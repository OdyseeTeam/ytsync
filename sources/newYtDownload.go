package sources

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/lbryio/ytsync/v5/downloader"
	"github.com/lbryio/ytsync/v5/ip_manager"
	"github.com/lbryio/ytsync/v5/timing"
	logUtils "github.com/lbryio/ytsync/v5/util"

	"github.com/lbryio/lbry.go/v2/extras/errors"
	"github.com/lbryio/lbry.go/v2/extras/stop"

	log "github.com/sirupsen/logrus"
)

type DownloadResults struct {
	CouldRetry       bool
	WasThrottled     bool
	ReduceResolution bool
	ChangeUserAgent  bool
	Successful       bool
	KnownError       error
}

func rawDownload(args []string, dir string) (*DownloadResults, error) {
	log.Printf("Running command yt-dlp %s", strings.Join(args, " "))
	cmd := exec.Command("yt-dlp", args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, errors.Err(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Err(err)
	}
	err = cmd.Start()
	if err != nil {
		return nil, errors.Err(err)
	}
	monitorStopGrp := stop.New()
	go detectSlowDownload(dir, monitorStopGrp, cmd)
	errorLog, _ := ioutil.ReadAll(stderr)
	outLog, _ := ioutil.ReadAll(stdout)
	err = cmd.Wait()
	monitorStopGrp.Stop()
	parsedFailure := parseFailureReason(string(errorLog))
	parsedOut := parseOutLog(string(outLog))
	if err != nil && !strings.Contains(err.Error(), "exit status 1") {
		return nil, errors.Err(err)
	}
	if errors.Unwrap(parsedFailure) != nil {
		dr := &DownloadResults{
			KnownError: parsedFailure,
		}
		switch errors.Unwrap(parsedFailure) {
		case ThrottledErr:
			dr.WasThrottled = true
			dr.CouldRetry = true
		case FragmentsRetriesErr, FormatNotAvailableErr:
			dr.CouldRetry = true
			dr.ReduceResolution = true
		}
		return dr, nil
	}
	if parsedOut != nil {
		dr := &DownloadResults{
			KnownError: parsedOut,
		}
		switch errors.Unwrap(parsedOut) {
		case ThrottledErr:
			dr.WasThrottled = true
			dr.CouldRetry = true
		case VideoTooLongErr:
			dr.CouldRetry = false
		case VideoTooBigErr:
			dr.CouldRetry = true
			dr.ReduceResolution = true
		}
		return dr, nil
	}

	return &DownloadResults{Successful: true}, nil
}

func detectSlowDownload(path string, stop *stop.Group, cmd *exec.Cmd) {
	stop.Add(1)
	defer stop.Done()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	count := 0
	lastSize := int64(0)
	for {
		select {
		case <-stop.Ch():
			return
		case <-ticker.C:
			size, err := logUtils.DirSize(path)
			if err != nil {
				log.Errorf("error while getting size of download directory: %s", errors.FullTrace(err))
				continue
			}
			delta := size - lastSize
			lastSize = size
			avgSpeed := delta / 10
			//log.Infof("download speed: %d bytes/s - min speed: %d", avgSpeed, 30*1024*1000)
			if avgSpeed < 30*1024 { //30 KB/s
				count++
			} else if count > 0 {
				count--
			}
			if count > 3 {
				err := cmd.Process.Signal(syscall.SIGKILL)
				if err != nil {
					log.Errorf("failure in killing slow download: %s", errors.Err(err))
					return
				}
				return
			}
		}
	}
}

var (
	ThrottledErr          = errors.Base("throttled")
	FragmentsRetriesErr   = errors.Base("missing fragments")
	VideoExtractionErr    = errors.Base("unable to extract")
	VideoTooLongErr       = errors.Base("video is too long to process")
	VideoTooBigErr        = errors.Base("the video is too big to sync, skipping for now")
	FormatNotAvailableErr = errors.Base("Requested format is not available")
)

const (
	ThrottledMsg          = "HTTP Error 429"
	AltThrottledMsg       = "returned non-zero exit status 8"
	FragmentsRetriesMsg   = "giving up after 0 fragment retries"
	UAVideoDataExtractMsg = "YouTube said: Unable to extract video data"
	DurationConstraintMsg = "does not pass filter (duration"
	SizeConstraintMsg     = "File is larger than max-filesize"
	FormatNotAvailable    = "Requested format is not available"
)

func parseFailureReason(errLog string) error {
	if errLog == "" {
		return nil
	}
	if strings.Contains(errLog, ThrottledMsg) {
		return errors.Err(ThrottledErr)
	}
	if strings.Contains(errLog, AltThrottledMsg) {
		return errors.Err(ThrottledErr)
	}
	if strings.Contains(errLog, FragmentsRetriesMsg) {
		return errors.Err(FragmentsRetriesErr)
	}
	if strings.Contains(errLog, UAVideoDataExtractMsg) {
		return errors.Err(VideoExtractionErr)
	}
	if strings.Contains(errLog, FormatNotAvailable) {
		return errors.Err(FormatNotAvailableErr)
	}
	return errors.Err(errLog)
}

func parseOutLog(outLog string) error {
	log.Debugln(outLog)
	if outLog == "" {
		return nil
	}
	if strings.Contains(outLog, DurationConstraintMsg) {
		return errors.Err(VideoTooLongErr)
	}
	if strings.Contains(outLog, SizeConstraintMsg) {
		return errors.Err(VideoTooBigErr)
	}
	if strings.Contains(outLog, ThrottledMsg) {
		return errors.Err(ThrottledErr)
	}
	return nil
}

func (v *YoutubeVideo) Xdownload() error {
	start := time.Now()
	defer func(start time.Time) {
		timing.TimedComponent("download").Add(time.Since(start))
	}(start)

	videoPath := v.getFullPath()

	err := os.Mkdir(v.videoDir(), 0777)
	if err != nil && !strings.Contains(err.Error(), "file exists") {
		return errors.Wrap(err, 0)
	}

	_, err = os.Stat(videoPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Err(err)
	} else if err == nil {
		log.Debugln(v.id + " already exists at " + videoPath)
		return nil
	}
	qualities := []string{
		"1080",
		"720",
		"480",
		"360",
	}
	dur := time.Duration(v.youtubeInfo.Duration) * time.Second
	if dur.Hours() > 1 { //for videos longer than 1 hour only sync up to 720p
		qualities = []string{
			"720",
			"480",
			"360",
		}
	}

	metadataPath := path.Join(logUtils.GetVideoMetadataDir(), v.id+".info.json")
	_, err = os.Stat(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Err("metadata information for video %s is missing! Why?", v.id)
		}
		return errors.Err(err)
	}

	metadata, err := parseVideoMetadata(metadataPath)

	err = checkCookiesIntegrity()
	if err != nil {
		return err
	}

	ytdlArgs := []string{
		fmt.Sprintf("https://www.youtube.com/watch?v=%s", v.id),
		"--no-progress",
		"-o" + strings.TrimSuffix(v.getFullPath(), ".mp4"),
		"--merge-output-format",
		"mp4",
		"--postprocessor-args",
		"ffmpeg:-movflags faststart",
		"--abort-on-unavailable-fragment",
		"--fragment-retries",
		"1",
		"--cookies",
		"cookies.txt",
		"--extractor-args",
		"youtube:player_client=android",
		"--load-info-json",
		metadataPath,
	}

	userAgent := []string{"--user-agent", downloader.ChromeUA}
	if v.maxVideoSize > 0 {
		ytdlArgs = append(ytdlArgs,
			"--max-filesize",
			fmt.Sprintf("%dM", v.maxVideoSize),
		)
	}
	if v.maxVideoLength > 0 {
		ytdlArgs = append(ytdlArgs,
			"--match-filter",
			fmt.Sprintf("duration <= %d", int(v.maxVideoLength.Seconds())),
		)
	}

	qualityIndex := 0
	remainingAttempts := 3

	for remainingAttempts > 0 {
		remainingAttempts--
		res, usedIp, err := func() (*DownloadResults, string, error) {
			sourceAddress, err := v.getSourceAddress()
			if err != nil {
				return nil, sourceAddress, err
			}
			defer v.pool.ReleaseIP(sourceAddress)
			quality := qualities[qualityIndex]
			dynamicArgs := append(ytdlArgs, "-fbestvideo[ext=mp4][vcodec!*=av01][height<="+quality+"]+bestaudio[ext!=webm][format_id!=258][format_id!=380][format_id!=251][format_id!=256][format_id!=327][format_id!=328]")
			dynamicArgs = append(dynamicArgs, userAgent...)
			dynamicArgs = append(dynamicArgs,
				"--source-address",
				sourceAddress,
			)
			dlStopGrp := stop.New()
			go v.trackProgressBar(dynamicArgs, metadata, dlStopGrp, sourceAddress)
			res, err := rawDownload(dynamicArgs, v.videoDir())
			//stop the progress bar
			dlStopGrp.Stop()
			return res, sourceAddress, err
		}()
		if err != nil {
			_ = v.delete(err.Error())
			return err
		}

		if res.Successful {
			fi, err := os.Stat(v.getFullPath())
			if err != nil {
				return errors.Err(err)
			}
			err = os.Chmod(v.getFullPath(), 0777)
			if err != nil {
				return errors.Err(err)
			}
			videoSize := fi.Size()
			v.size = &videoSize
			return nil
		}
		if res.CouldRetry {
			if res.ReduceResolution {
				qualityIndex++
				_ = v.delete(res.KnownError.Error())
				if qualityIndex >= len(qualities)-1 {
					//return errors.Err("could not find a suitable resolution")
					break //todo: can this ever happen?
				}
				continue
			}
			if res.ChangeUserAgent {
				log.Infof("trying different user agent for video %s", v.ID())
				userAgent = []string{"--user-agent", downloader.GoogleBotUA}
				continue
			}
			if res.WasThrottled {
				remainingAttempts++
				v.pool.SetThrottled(usedIp)
				continue
			}
		}
		_ = v.delete(res.KnownError.Error())
		return res.KnownError
	}
	return nil
}

func (v *YoutubeVideo) getSourceAddress() (string, error) {
	for {
		sourceAddress, err := v.pool.GetIP(v.id)
		if err == nil {
			return sourceAddress, nil
		}
		if errors.Is(err, ip_manager.ErrAllThrottled) {
			select {
			case <-v.stopGroup.Ch():
				return sourceAddress, errors.Err("interrupted by user")
			default:
				time.Sleep(ip_manager.IPCooldownPeriod)
				continue
			}
		}
		return sourceAddress, err
	}
}
