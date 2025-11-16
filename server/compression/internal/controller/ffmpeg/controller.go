package ffmpeg

import (
	"context"
	"encoding/json"
	"errors"
	"ffmpeg/wrapper/compression/internal/repository"
	compressionModel "ffmpeg/wrapper/compression/pkg/model"
	metadataModel "ffmpeg/wrapper/metadata/pkg/model"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/segmentio/kafka-go"
)

var ErrNotFound = errors.New("not found")

type Controller struct {
	kafkaReader *kafka.Reader
	kafkaWriter *kafka.Writer
	repo        repository.S3
}

func New(reader *kafka.Reader, writer *kafka.Writer, repository repository.S3) *Controller {
	return &Controller{
		kafkaReader: reader,
		kafkaWriter: writer,
		repo:        repository,
	}
}

var bucketName = os.Getenv("bucketname")

func (c *Controller) Compress(ctx context.Context, duration float64, compressedKey string, objectKey string, filename string) (*v4.PresignedHTTPRequest, error) {
	target_video := 8.0 //megabytes
	target_audio := 1.0 // megabytes
	// get the video from db here
	videoBitrate, audioBitrate := CalculateBitrates(duration, target_video, target_audio)

	videoBitrateStr := strconv.FormatFloat(videoBitrate, 'f', 0, 64)
	audioBitrateStr := strconv.FormatFloat(audioBitrate, 'f', 0, 64)

	filePath, err := c.repo.DownloadObject(ctx, bucketName, objectKey, filename)
	if err != nil {
		return nil, fmt.Errorf("error downloading object from R2: %w", err)
	}
	outputFilename := "/tmp/" + fmt.Sprintf("compressed_%s.mp4", filename)
	// PASS 1
	log.Println(outputFilename, filePath, filename)
	cmd1 := exec.Command(
		"ffmpeg",
		"-y",
		"-i", filePath,
		"-c:v", "libx265",
		"-preset", "medium",
		"-b:v", videoBitrateStr,
		"-pass", "1", "-passlogfile", "/tmp/passlog",
		"-c:a", "aac",
		"-b:a", audioBitrateStr,
		"-f", "mp4", "/dev/null",
	)
	cmd1.Dir = "/tmp"
	cmd1.Stderr = os.Stderr
	cmd1.Stdout = os.Stdout
	if err := cmd1.Run(); err != nil {
		return nil, fmt.Errorf("error running ffmpeg pass 1 %w", err)
	}

	// PASS 2

	cmd2 := exec.Command(
		"ffmpeg",
		"-i", filePath,
		"-c:v", "libx265",
		"-preset", "medium",
		"-b:v", videoBitrateStr,
		"-pass", "2", "-passlogfile", "/tmp/passlog",
		"-c:a", "aac",
		"-b:a", audioBitrateStr,
		outputFilename,
	)
	cmd2.Dir = "/tmp"
	cmd2.Stderr = os.Stderr
	cmd2.Stdout = os.Stdout
	if err := cmd2.Run(); err != nil {
		return nil, fmt.Errorf("error running ffmpeg pass 2 %w", err)
	}

	os.Remove("/tmp/passlog-0.log")
	os.Remove("/tmp/passlog-0.log.mbtree")

	err = c.repo.UploadObject(ctx, bucketName, compressedKey, outputFilename)
	if err != nil {
		return nil, fmt.Errorf("error uploading object to R2: %w", err)
	}
	presignedRequest, err := c.repo.GetObject(ctx, bucketName, compressedKey, 120)
	if err != nil {
		return nil, fmt.Errorf("failed to create presigned download url: %w", err)
	}
	os.Remove(outputFilename)
	return presignedRequest, nil
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

func (c *Controller) ConsumeCompressionEvent(ctx context.Context) {

	c.kafkaReader.SetOffset(kafka.LastOffset)
	for {
		m, err := c.kafkaReader.ReadMessage(ctx)
		if err != nil {
			break
		}

		var event metadataModel.CompressionEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("Unmarshal error: %v", err)
			continue
		}

		// Check if this is actually a CompressionEvent with metadata
		if event.Metadata.Duration == "" {
			log.Printf("Skipping message without duration (not a CompressionEvent?)")
			continue
		}
		durationFloat, err := strconv.ParseFloat(event.Metadata.Duration, 64)
		if err != nil {
			log.Printf("failed to parse duration: %v", err)
			// handle error or skip processing
			continue
		}

		compressedKey := fmt.Sprintf("%s_compressed", event.ObjectKey)
		presignedDownloadURL, err := c.Compress(ctx, durationFloat, compressedKey, event.ObjectKey, event.ObjectKey)

		if err != nil {
			log.Printf("compression failed: %v", err)
			_ = c.PublishCompressionResultEvent(ctx, compressionModel.CompressionEventTypeFail, event.JobID, event.ObjectKey, compressedKey, nil)
			continue
		}

		err = c.PublishCompressionResultEvent(ctx, compressionModel.CompressionEventTypeSuccess, event.JobID, event.ObjectKey, compressedKey, presignedDownloadURL)
		if err != nil {
			log.Printf("failed to publish compression result: %v", err)
		}
	}

	if err := c.kafkaReader.Close(); err != nil {
		log.Fatal("failed to close reader", err)
	}
}

func (c *Controller) PublishCompressionResultEvent(ctx context.Context, eventType compressionModel.CompressionEventType,
	jobID int64, objecKey string, compressedKey string, presignedDownloadURL *v4.PresignedHTTPRequest) error {

	var presignedPayload *compressionModel.PresignedRequestPayload
	if presignedDownloadURL != nil {
		headers := make(map[string]string)
		for k, v := range presignedDownloadURL.SignedHeader {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		presignedPayload = &compressionModel.PresignedRequestPayload{
			URL:     presignedDownloadURL.URL,
			Method:  presignedDownloadURL.Method,
			Headers: headers,
		}
	}
	event := compressionModel.CompressionResultEvent{
		CompressionEventType: eventType,
		JobID:                jobID,
		ObjectKey:            objecKey,
		CompressedKey:        compressedKey,
		PresignedDownloadUrl: presignedPayload,
	}

	if eventType == compressionModel.CompressionEventTypeFail {
		event.PresignedDownloadUrl = nil
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal compression event: %w", err)
	}
	return c.kafkaWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("%d", jobID)),
		Value: payload,
	})
}
