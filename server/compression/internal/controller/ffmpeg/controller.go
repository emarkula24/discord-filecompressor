package ffmpeg

import (
	"context"
	"errors"
	"ffmpeg/wrapper/compression/pkg/model"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
)

var ErrNotFound = errors.New("not found")

type Controller struct {
}

func New() *Controller {
	return &Controller{}
}

func (c *Controller) Compress(ctx context.Context, duration float64, input string) (*model.CompressedLink, error) {
	target_video := 8.3 //megabytes
	target_audio := 1.2 // megabytes
	// get the video from db here
	videoBitrate, audioBitrate := CalculateBitrates(duration, target_video, target_audio)
	output := RandStringBytes(20)

	videoBitrateStr := strconv.FormatFloat(videoBitrate, 'f', 0, 64)
	audioBitrateStr := strconv.FormatFloat(audioBitrate, 'f', 0, 64)

	// PASS 1
	cmd1 := exec.Command(
		"ffmpeg",
		"-y",
		"-i", input,
		"-c:v", "libx264",
		"-preset", "medium",
		"-b:v", videoBitrateStr,
		"-pass", "1",
		"-c:a", "aac",
		"-b:a", audioBitrateStr,
		"-f", "mp4", "/dev/null",
	)

	cmd1.Stderr = os.Stderr
	if err := cmd1.Run(); err != nil {
		return nil, fmt.Errorf("error running ffmpeg pass 1 %w", err)
	}

	// PASS 2
	cmd2 := exec.Command(
		"ffmpeg",
		"-i", input,
		"-c:v", "libx264",
		"-preset", "medium",
		"-b:v", videoBitrateStr,
		"-pass", "2",
		"-c:a", "aac",
		"-b:a", audioBitrateStr,
		output,
	)

	cmd2.Stderr = os.Stderr
	if err := cmd2.Run(); err != nil {
		return nil, fmt.Errorf("error running ffmpeg pass 2 %w", err)
	}
	// upload the video here, return link to it if possible
	v := "link to the uploaded videofile."
	a := model.CompressedLink(v)
	return &a, nil
}

func CalculateBitrates(duration float64, targetVideo, targetAudio float64) (float64, float64) {

	targetVideoBitrate := float64(targetVideo) * float64(8388.608) / duration
	targetAudioBitrate := float64(targetAudio) * float64(8388.608) / duration
	// result is in kb, ffmpeg requires bytes, thus the conversion here.
	return targetVideoBitrate * 1000, targetAudioBitrate * 1000
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
