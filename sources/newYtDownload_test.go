package sources

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/lbryio/lbry.go/v2/extras/errors"

	"github.com/stretchr/testify/assert"
)

func Test_rawDownload(t *testing.T) {
	nowString := fmt.Sprintf("%d", time.Now().Unix())
	testPath, err := os.MkdirTemp("/tmp/", "")
	if !assert.NoError(t, err) {
		return
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Error(err)
		}
	}(testPath)
	args := []string{
		"--merge-output-format",
		"mp4",
		"-o" + fmt.Sprintf("%s/%s.mp4", testPath, nowString),
		"--postprocessor-args",
		"ffmpeg:-movflags faststart",
		"--abort-on-unavailable-fragment",
		"--fragment-retries",
		"1",
		"--extractor-args",
		"youtube:player_client=android",
		"--max-filesize", "3050M",
		"--match-filter", "duration <= 7200",
		"-f", "bestvideo[ext=mp4][vcodec!*=av01][height<=1920]+bestaudio[ext!=webm][format_id!=258][format_id!=380][format_id!=251][format_id!=256][format_id!=327][format_id!=328]",
		"--user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
		"https://www.youtube.com/watch?v=HYH4Z__jqe0",
	}
	res, err := rawDownload(args, testPath)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, res) {
		return
	}
	if !assert.NoError(t, res.KnownError) {
		return
	}
}

func TestVideoTooLong(t *testing.T) {
	nowString := fmt.Sprintf("%d", time.Now().Unix())
	testPath, err := os.MkdirTemp("/tmp/", "")
	if !assert.NoError(t, err) {
		return
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Error(err)
		}
	}(testPath)
	args := []string{
		"--merge-output-format",
		"mp4",
		"-o" + fmt.Sprintf("%s/%s.mp4", testPath, nowString),
		"--postprocessor-args",
		"ffmpeg:-movflags faststart",
		"--abort-on-unavailable-fragment",
		"--fragment-retries",
		"1",
		"--extractor-args",
		"youtube:player_client=android",
		"--max-filesize", "3050M",
		"--match-filter", "duration <= 7200",
		"-f", "bestvideo[ext=mp4][vcodec!*=av01][height<=1920]+bestaudio[ext!=webm][format_id!=258][format_id!=380][format_id!=251][format_id!=256][format_id!=327][format_id!=328]",
		"--user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
		"https://www.youtube.com/watch?v=X0RK2jz5HOI",
	}
	res, err := rawDownload(args, testPath)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, res) {
		return
	}
	assert.True(t, errors.Is(res.KnownError, VideoTooLongErr))
}

func TestTooBig(t *testing.T) {
	nowString := fmt.Sprintf("%d", time.Now().Unix())
	testPath, err := os.MkdirTemp("/tmp/", "")
	if !assert.NoError(t, err) {
		return
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Error(err)
		}
	}(testPath)
	args := []string{
		"--merge-output-format",
		"mp4",
		"-o" + fmt.Sprintf("%s/%s.mp4", testPath, nowString),
		"--postprocessor-args",
		"ffmpeg:-movflags faststart",
		"--abort-on-unavailable-fragment",
		"--fragment-retries",
		"1",
		"--extractor-args",
		"youtube:player_client=android",
		"--max-filesize", "1M",
		"--match-filter", "duration <= 7200",
		"-f", "bestvideo[ext=mp4][vcodec!*=av01][height<=1920]+bestaudio[ext!=webm][format_id!=258][format_id!=380][format_id!=251][format_id!=256][format_id!=327][format_id!=328]",
		"--user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
		"https://www.youtube.com/watch?v=HYH4Z__jqe0",
	}
	res, err := rawDownload(args, testPath)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, res) {
		return
	}
	assert.True(t, errors.Is(res.KnownError, VideoTooBigErr))
}

func TestNoFormat(t *testing.T) {
	nowString := fmt.Sprintf("%d", time.Now().Unix())
	testPath, err := os.MkdirTemp("/tmp/", "")
	if !assert.NoError(t, err) {
		return
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Error(err)
		}
	}(testPath)
	args := []string{
		"--merge-output-format",
		"mp4",
		"-o" + fmt.Sprintf("%s/%s.mp4", testPath, nowString),
		"--postprocessor-args",
		"ffmpeg:-movflags faststart",
		"--abort-on-unavailable-fragment",
		"--fragment-retries",
		"1",
		"--extractor-args",
		"youtube:player_client=android",
		"--max-filesize", "1M",
		"--match-filter", "duration <= 7200",
		"-f", "bestvideo[ext=mp4][vcodec!*=av01][height<=1080]+bestaudio[ext!=webm][format_id!=258][format_id!=380][format_id!=251][format_id!=256][format_id!=327][format_id!=328]",
		"--user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
		"https://www.youtube.com/watch?v=w92FBmzJnc4",
	}
	res, err := rawDownload(args, testPath)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, res) {
		return
	}
	assert.True(t, errors.Is(res.KnownError, FormatNotAvailableErr))
	assert.True(t, res.CouldRetry)
	assert.True(t, res.ReduceResolution)
}
