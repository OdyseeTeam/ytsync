package downloader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/lbryio/ytsync/v5/downloader/ytdl"
	"github.com/lbryio/ytsync/v5/ip_manager"
	"github.com/lbryio/ytsync/v5/shared"
	util2 "github.com/lbryio/ytsync/v5/util"

	"github.com/lbryio/lbry.go/v2/extras/errors"
	"github.com/lbryio/lbry.go/v2/extras/stop"
	"github.com/lbryio/lbry.go/v2/extras/util"

	"github.com/sirupsen/logrus"
)

func GetPlaylistVideoIDs(channelName string, maxVideos int, stopChan stop.Chan, pool *ip_manager.IPPool) ([]string, error) {
	tabs := []string{"videos", "shorts", "streams"}
	var videoIDs []string
	for _, tab := range tabs {
		args := []string{"--skip-download", "https://www.youtube.com/channel/" + channelName + "/" + tab, "--get-id", "--flat-playlist", "--cookies", "cookies.txt", "--playlist-end", fmt.Sprintf("%d", maxVideos)}
		ids, err := run(channelName, args, stopChan, pool)
		if err != nil {
			if strings.Contains(err.Error(), "This channel does not have a") {
				continue
			}
			if strings.Contains(err.Error(), "Incomplete data received") {
				continue
			}
			return nil, errors.Err(err)
		}
		for i, v := range ids {
			if v == "" {
				continue
			}
			if i >= maxVideos {
				break
			}
			videoIDs = append(videoIDs, v)
		}
	}
	return videoIDs, nil
}

const releaseTimeFormat = "2006-01-02, 15:04:05 (MST)"

func GetVideoInformation(videoID string, stopChan stop.Chan, pool *ip_manager.IPPool) (*ytdl.YtdlVideo, error) {
	args := []string{
		"--skip-download",
		"--write-info-json",
		fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
		"--cookies",
		"cookies.txt",
		"-o",
		path.Join(util2.GetVideoMetadataDir(), videoID),
	}
	_, err := run(videoID, args, stopChan, pool)
	if err != nil {
		return nil, errors.Err(err)
	}

	f, err := os.Open(path.Join(util2.GetVideoMetadataDir(), videoID+".info.json"))
	if err != nil {
		return nil, errors.Err(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer f.Close()
	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(f)

	var video *ytdl.YtdlVideo
	err = json.Unmarshal(byteValue, &video)
	if err != nil {
		return nil, errors.Err(err)
	}

	return video, nil
}

const (
	GoogleBotUA             = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
	ChromeUA                = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
	maxAttempts             = 3
	extractionError         = "YouTube said: Unable to extract video data"
	throttledError          = "HTTP Error 429"
	AlternateThrottledError = "returned non-zero exit status 8"
	youtubeDlError          = "exit status 1"
	videoPremiereError      = "Premieres in"
	liveEventError          = "This live event will begin in"
)

func run(use string, args []string, stopChan stop.Chan, pool *ip_manager.IPPool) ([]string, error) {
	var useragent []string
	var lastError error
	for attempts := 0; attempts < maxAttempts; attempts++ {
		sourceAddress, err := getIPFromPool(use, stopChan, pool)
		if err != nil {
			return nil, err
		}
		argsForCommand := append(args, "--source-address", sourceAddress)
		argsForCommand = append(argsForCommand, useragent...)
		binary := "yt-dlp"
		cmd := exec.Command(binary, argsForCommand...)

		res, err := runCmd(cmd, stopChan)
		pool.ReleaseIP(sourceAddress)
		if err == nil {
			return res, nil
		}
		lastError = err
		if strings.Contains(err.Error(), youtubeDlError) {
			if util.SubstringInSlice(err.Error(), shared.ErrorsNoRetry) {
				break
			}
			if strings.Contains(err.Error(), extractionError) {
				logrus.Warnf("known extraction error: %s", errors.FullTrace(err))
				useragent = nextUA(useragent)
			}
			if strings.Contains(err.Error(), throttledError) || strings.Contains(err.Error(), AlternateThrottledError) {
				pool.SetThrottled(sourceAddress)
				//we don't want throttle errors to count toward the max retries
				attempts--
			}
		}
	}
	return nil, lastError
}

func nextUA(current []string) []string {
	if len(current) == 0 {
		return []string{"--user-agent", GoogleBotUA}
	}
	return []string{"--user-agent", ChromeUA}
}

func runCmd(cmd *exec.Cmd, stopChan stop.Chan) ([]string, error) {
	logrus.Infof("running yt-dlp cmd: %s", strings.Join(cmd.Args, " "))
	var err error
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
	outLog, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, errors.Err(err)
	}
	errorLog, err := ioutil.ReadAll(stderr)
	if err != nil {
		return nil, errors.Err(err)
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-stopChan:
		err := cmd.Process.Kill()
		if err != nil {
			return nil, errors.Prefix("failed to kill command after stopper cancellation", err)
		}
		return nil, errors.Err("interrupted by user")
	case err := <-done:
		if err != nil {
			//return nil, errors.Prefix("yt-dlp "+strings.Join(cmd.Args, " ")+" ["+string(errorLog)+"]", err)
			return nil, errors.Prefix(string(errorLog), err)
		}
		return strings.Split(strings.Replace(string(outLog), "\r\n", "\n", -1), "\n"), nil
	}
}

func getIPFromPool(use string, stopChan stop.Chan, pool *ip_manager.IPPool) (sourceAddress string, err error) {
	for {
		sourceAddress, err = pool.GetIP(use)
		if err != nil {
			if errors.Is(err, ip_manager.ErrAllThrottled) {
				select {
				case <-stopChan:
					return "", errors.Err("interrupted by user")

				default:
					time.Sleep(ip_manager.IPCooldownPeriod)
					continue
				}
			} else {
				return "", err
			}
		}
		break
	}
	return
}
