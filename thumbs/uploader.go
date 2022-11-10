package thumbs

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/lbryio/ytsync/v5/configs"
	"github.com/lbryio/ytsync/v5/downloader/ytdl"

	"github.com/lbryio/lbry.go/v2/extras/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	log "github.com/sirupsen/logrus"
)

type thumbnailUploader struct {
	name        string
	originalUrl string
	mirroredUrl string
	s3Config    aws.Config
}

const thumbnailPath = "/tmp/ytsync_thumbnails/"
const ThumbnailEndpoint = "https://thumbnails.lbry.com/"
const EmptyThumbnailHash = "e2ddfee11ae7edcae257da47f3a78a70"

func (u *thumbnailUploader) downloadThumbnail() error {
	_ = os.Mkdir(thumbnailPath, 0777)
	img, err := os.Create("/tmp/ytsync_thumbnails/" + u.name)
	if err != nil {
		return errors.Err(err)
	}
	defer img.Close()
	if strings.HasPrefix(u.originalUrl, "//") {
		u.originalUrl = "https:" + u.originalUrl
	}
	resp, err := http.Get(u.originalUrl)
	if err != nil {
		return errors.Err(err)
	}
	defer resp.Body.Close()

	reader, thumbnailHash, err := getMD5(resp.Body)
	if err != nil {
		return errors.Err(err)
	}
	if thumbnailHash == EmptyThumbnailHash {
		return errors.Err("youtube thumbnail is empty - too soon?")
	}
	_, err = io.Copy(img, reader)
	if err != nil {
		return errors.Err(err)
	}
	return nil
}

func getMD5(body io.Reader) (io.Reader, string, error) {
	var b bytes.Buffer

	hash := md5.New()
	_, err := io.Copy(&b, io.TeeReader(body, hash))

	if err != nil {
		return nil, "", errors.Err(err)
	}

	return &b, hex.EncodeToString(hash.Sum(nil)), nil
}

func (u *thumbnailUploader) uploadThumbnail() error {
	key := &u.name
	thumb, err := os.Open("/tmp/ytsync_thumbnails/" + u.name)
	if err != nil {
		return errors.Err(err)
	}
	defer thumb.Close()

	s3Session, err := session.NewSession(&u.s3Config)
	if err != nil {
		return errors.Err(err)
	}

	uploader := s3manager.NewUploader(s3Session)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:       aws.String("thumbnails.lbry.com"),
		Key:          key,
		Body:         thumb,
		ACL:          aws.String("public-read"),
		ContentType:  aws.String("image/jpeg"),
		CacheControl: aws.String("public, max-age=2592000"),
	})

	u.mirroredUrl = ThumbnailEndpoint + u.name
	return errors.Err(err)
}

func (u *thumbnailUploader) deleteTmpFile() {
	err := os.Remove("/tmp/ytsync_thumbnails/" + u.name)
	if err != nil {
		log.Infof("failed to delete local thumbnail file: %s", err.Error())
	}
}
func MirrorThumbnail(url string, name string) (string, error) {
	tu := thumbnailUploader{
		originalUrl: url,
		name:        name,
		s3Config:    *configs.Configuration.AWSThumbnailsS3Config.GetS3AWSConfig(),
	}
	err := tu.downloadThumbnail()
	if err != nil {
		return "", err
	}
	defer tu.deleteTmpFile()

	err = tu.uploadThumbnail()
	if err != nil {
		return "", err
	}

	//this is our own S3 storage
	tu2 := thumbnailUploader{
		originalUrl: url,
		name:        name,
		s3Config:    *configs.Configuration.ThumbnailsS3Config.GetS3AWSConfig(),
	}
	err = tu2.uploadThumbnail()
	if err != nil {
		return "", err
	}

	return tu.mirroredUrl, nil
}

// GetBestThumbnail returns the thumbnail URL for a given video ID
// youtube keeps changing how their thumbnails are listed
// as of october 2022 this method seems to be always returning a valid thumbnail
func GetBestThumbnail(thumbnails []ytdl.Thumbnail) *ytdl.Thumbnail {
	var bestWidth ytdl.Thumbnail
	for _, thumbnail := range thumbnails {
		if bestWidth.Width < thumbnail.Width {
			bestWidth = thumbnail
		}
	}
	return &bestWidth
}
