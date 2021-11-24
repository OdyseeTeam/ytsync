package ytapi

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lbryio/ytsync/v5/shared"
	logUtils "github.com/lbryio/ytsync/v5/util"
	"github.com/vbauerster/mpb/v7"

	"github.com/lbryio/ytsync/v5/downloader/ytdl"

	"github.com/lbryio/ytsync/v5/downloader"
	"github.com/lbryio/ytsync/v5/ip_manager"
	"github.com/lbryio/ytsync/v5/sdk"
	"github.com/lbryio/ytsync/v5/sources"

	"github.com/lbryio/lbry.go/v2/extras/errors"
	"github.com/lbryio/lbry.go/v2/extras/jsonrpc"
	"github.com/lbryio/lbry.go/v2/extras/stop"
	"github.com/lbryio/lbry.go/v2/extras/util"

	log "github.com/sirupsen/logrus"
)

type Video interface {
	Size() *int64
	ID() string
	IDAndNum() string
	PlaylistPosition() int
	PublishedAt() time.Time
	Sync(*jsonrpc.Client, sources.SyncParams, *sdk.SyncedVideo, bool, *sync.RWMutex, *sync.WaitGroup, *mpb.Progress) (*sources.SyncSummary, error)
}

type byPublishedAt []Video

func (a byPublishedAt) Len() int           { return len(a) }
func (a byPublishedAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPublishedAt) Less(i, j int) bool { return a[i].PublishedAt().Before(a[j].PublishedAt()) }

type VideoParams struct {
	VideoDir string
	Stopper  *stop.Group
	IPPool   *ip_manager.IPPool
}

var mostRecentlyFailedChannel string // TODO: fix this hack!

func GetVideosToSync(config *sdk.APIConfig, channelID string, syncedVideos map[string]sdk.SyncedVideo, quickSync bool, maxVideos int, videoParams VideoParams, lastUploadedVideo string) ([]Video, error) {
	var videos []Video
	if quickSync && maxVideos > 50 {
		maxVideos = 50
	}
	allVideos, err := downloader.GetPlaylistVideoIDs(channelID, maxVideos, videoParams.Stopper.Ch(), videoParams.IPPool)
	if err != nil {
		return nil, errors.Err(err)
	}
	videoIDs := make([]string, 0, len(allVideos))
	for _, video := range allVideos {
		sv, ok := syncedVideos[video]
		if ok && util.SubstringInSlice(sv.FailureReason, shared.NeverRetryFailures) {
			continue
		}
		videoIDs = append(videoIDs, video)
	}
	log.Infof("Got info for %d videos from youtube downloader", len(videoIDs))

	playlistMap := make(map[string]int64)
	for i, videoID := range videoIDs {
		playlistMap[videoID] = int64(i)
	}
	//this will ensure that we at least try to sync the video that was marked as last uploaded video in the database.
	if lastUploadedVideo != "" {
		_, ok := playlistMap[lastUploadedVideo]
		if !ok {
			playlistMap[lastUploadedVideo] = 0
			videoIDs = append(videoIDs, lastUploadedVideo)
		}
	}

	if len(videoIDs) < 1 {
		if channelID == mostRecentlyFailedChannel {
			return nil, errors.Err("playlist items not found")
		}
		mostRecentlyFailedChannel = channelID
	}

	vids, err := getVideos(config, channelID, videoIDs, videoParams.Stopper.Ch(), videoParams.IPPool)
	if err != nil {
		return nil, err
	}

	for _, item := range vids {
		positionInList := playlistMap[item.ID]
		videoToAdd, err := sources.NewYoutubeVideo(videoParams.VideoDir, item, positionInList, videoParams.Stopper, videoParams.IPPool)
		if err != nil {
			return nil, errors.Err(err)
		}
		videos = append(videos, videoToAdd)
	}

	for k, v := range syncedVideos {
		if !v.Published {
			continue
		}
		if _, ok := playlistMap[k]; !ok {
			videos = append(videos, sources.NewMockedVideo(videoParams.VideoDir, k, channelID, videoParams.Stopper, videoParams.IPPool))
		}
	}

	sort.Sort(byPublishedAt(videos))

	return videos, nil
}

// CountVideosInChannel is unused for now... keeping it here just in case
func CountVideosInChannel(channelID string) (int, error) {
	url := "https://socialblade.com/youtube/channel/" + channelID

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("User-Agent", downloader.ChromeUA)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "socialblade.com")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, errors.Err(err)
	}
	defer res.Body.Close()

	var line string
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "youtube-stats-header-uploads") {
			line = scanner.Text()
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}
	if line == "" {
		return 0, errors.Err("upload count line not found")
	}

	matches := regexp.MustCompile(">([0-9]+)<").FindStringSubmatch(line)
	if len(matches) != 2 {
		return 0, errors.Err("upload count not found with regex")
	}

	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, errors.Err(err)
	}

	return num, nil
}

func ChannelInfo(channelID string) (*YoutubeStatsResponse, error) {
	url := "https://www.youtube.com/channel/" + channelID + "/about"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("User-Agent", downloader.ChromeUA)
	req.Header.Add("Accept", "*/*")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Err(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Err(err)
	}
	pageBody := string(body)
	dataStartIndex := strings.Index(pageBody, "window[\"ytInitialData\"] = ") + 26
	if dataStartIndex == 25 {
		dataStartIndex = strings.Index(pageBody, "var ytInitialData = ") + 20
	}
	dataEndIndex := strings.Index(pageBody, "]}}};") + 4
	if dataEndIndex < dataStartIndex {
		return nil, errors.Err("start index is lower than end index. cannot extract channel info!")
	}
	data := pageBody[dataStartIndex:dataEndIndex]
	var decodedResponse YoutubeStatsResponse
	err = json.Unmarshal([]byte(data), &decodedResponse)
	if err != nil {
		return nil, errors.Err(err)
	}

	return &decodedResponse, nil
}

func getVideos(config *sdk.APIConfig, channelID string, videoIDs []string, stopChan stop.Chan, ipPool *ip_manager.IPPool) ([]*ytdl.YtdlVideo, error) {
	var videos []*ytdl.YtdlVideo
	for _, videoID := range videoIDs {
		if len(videoID) < 5 {
			continue
		}
		select {
		case <-stopChan:
			return videos, errors.Err("interrupted by user")
		default:
		}

		//ip, err := ipPool.GetIP(videoID)
		//if err != nil {
		//	return nil, err
		//}
		//video, err := downloader.GetVideoInformation(videoID, &net.TCPAddr{IP: net.ParseIP(ip)})
		state, err := config.VideoState(videoID)
		if err != nil {
			return nil, errors.Err(err)
		}
		if state == "published" {
			continue
		}
		video, err := downloader.GetVideoInformation(config, videoID, stopChan, nil, ipPool)
		if err != nil {
			errSDK := config.MarkVideoStatus(shared.VideoStatus{
				ChannelID:     channelID,
				VideoID:       videoID,
				Status:        "failed",
				FailureReason: err.Error(),
			})
			logUtils.SendErrorToSlack(fmt.Sprintf("Skipping video (%s): %s", videoID, errors.FullTrace(err)))
			if errSDK != nil {
				return nil, errors.Err(errSDK)
			}
		} else {
			videos = append(videos, video)
		}
	}
	return videos, nil
}

type YoutubeStatsResponse struct {
	Contents struct {
		TwoColumnBrowseResultsRenderer struct {
			Tabs []struct {
				TabRenderer struct {
					Title    string `json:"title"`
					Selected bool   `json:"selected"`
					Content  struct {
						SectionListRenderer struct {
							Contents []struct {
								ItemSectionRenderer struct {
									Contents []struct {
										ChannelAboutFullMetadataRenderer struct {
											Description struct {
												SimpleText string `json:"simpleText"`
											} `json:"description"`
											ViewCountText struct {
												SimpleText string `json:"simpleText"`
											} `json:"viewCountText"`
											JoinedDateText struct {
												Runs []struct {
													Text string `json:"text"`
												} `json:"runs"`
											} `json:"joinedDateText"`
											CanonicalChannelURL        string `json:"canonicalChannelUrl"`
											BypassBusinessEmailCaptcha bool   `json:"bypassBusinessEmailCaptcha"`
											Title                      struct {
												SimpleText string `json:"simpleText"`
											} `json:"title"`
											Avatar struct {
												Thumbnails []struct {
													URL    string `json:"url"`
													Width  int    `json:"width"`
													Height int    `json:"height"`
												} `json:"thumbnails"`
											} `json:"avatar"`
											ShowDescription  bool `json:"showDescription"`
											DescriptionLabel struct {
												Runs []struct {
													Text string `json:"text"`
												} `json:"runs"`
											} `json:"descriptionLabel"`
											DetailsLabel struct {
												Runs []struct {
													Text string `json:"text"`
												} `json:"runs"`
											} `json:"detailsLabel"`
											ChannelID string `json:"channelId"`
										} `json:"channelAboutFullMetadataRenderer"`
									} `json:"contents"`
								} `json:"itemSectionRenderer"`
							} `json:"contents"`
						} `json:"sectionListRenderer"`
					} `json:"content"`
				} `json:"tabRenderer"`
			} `json:"tabs"`
		} `json:"twoColumnBrowseResultsRenderer"`
	} `json:"contents"`
	Header struct {
		C4TabbedHeaderRenderer struct {
			ChannelID string `json:"channelId"`
			Title     string `json:"title"`
			Avatar    struct {
				Thumbnails []struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"thumbnails"`
			} `json:"avatar"`
			Banner struct {
				Thumbnails []struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"thumbnails"`
			} `json:"banner"`
			VisitTracking struct {
				RemarketingPing string `json:"remarketingPing"`
			} `json:"visitTracking"`
			SubscriberCountText struct {
				SimpleText string `json:"simpleText"`
			} `json:"subscriberCountText"`
		} `json:"c4TabbedHeaderRenderer"`
	} `json:"header"`
	Metadata struct {
		ChannelMetadataRenderer struct {
			Title                string   `json:"title"`
			Description          string   `json:"description"`
			RssURL               string   `json:"rssUrl"`
			ChannelConversionURL string   `json:"channelConversionUrl"`
			ExternalID           string   `json:"externalId"`
			Keywords             string   `json:"keywords"`
			OwnerUrls            []string `json:"ownerUrls"`
			Avatar               struct {
				Thumbnails []struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"thumbnails"`
			} `json:"avatar"`
			ChannelURL       string `json:"channelUrl"`
			IsFamilySafe     bool   `json:"isFamilySafe"`
			VanityChannelURL string `json:"vanityChannelUrl"`
		} `json:"channelMetadataRenderer"`
	} `json:"metadata"`
	Topbar struct {
		DesktopTopbarRenderer struct {
			CountryCode string `json:"countryCode"`
		} `json:"desktopTopbarRenderer"`
	} `json:"topbar"`
	Microformat struct {
		MicroformatDataRenderer struct {
			URLCanonical string `json:"urlCanonical"`
			Title        string `json:"title"`
			Description  string `json:"description"`
			Thumbnail    struct {
				Thumbnails []struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"thumbnails"`
			} `json:"thumbnail"`
			SiteName           string   `json:"siteName"`
			AppName            string   `json:"appName"`
			AndroidPackage     string   `json:"androidPackage"`
			IosAppStoreID      string   `json:"iosAppStoreId"`
			IosAppArguments    string   `json:"iosAppArguments"`
			OgType             string   `json:"ogType"`
			URLApplinksWeb     string   `json:"urlApplinksWeb"`
			URLApplinksIos     string   `json:"urlApplinksIos"`
			URLApplinksAndroid string   `json:"urlApplinksAndroid"`
			URLTwitterIos      string   `json:"urlTwitterIos"`
			URLTwitterAndroid  string   `json:"urlTwitterAndroid"`
			TwitterCardType    string   `json:"twitterCardType"`
			TwitterSiteHandle  string   `json:"twitterSiteHandle"`
			SchemaDotOrgType   string   `json:"schemaDotOrgType"`
			Noindex            bool     `json:"noindex"`
			Unlisted           bool     `json:"unlisted"`
			FamilySafe         bool     `json:"familySafe"`
			Tags               []string `json:"tags"`
		} `json:"microformatDataRenderer"`
	} `json:"microformat"`
}
