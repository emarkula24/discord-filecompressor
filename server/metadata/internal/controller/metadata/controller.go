package metadata

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"ffmpeg/wrapper/metadata/internal/repository"
	"ffmpeg/wrapper/metadata/pkg/model"
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
	"gopkg.in/vansante/go-ffprobe.v2"
)

var bucketName = os.Getenv("bucketname")
var ErrNotFound = errors.New("not found")

type Controller struct {
	repo        repository.S3
	kafkaWriter *kafka.Writer
}

func New(repository repository.S3, writer *kafka.Writer) *Controller {
	return &Controller{
		repo:        repository,
		kafkaWriter: writer,
	}
}

func (c *Controller) GetMetadata(ctx context.Context, objectKey string) (*model.Metadata, error) {
	filename, err := c.repo.DownloadObject(ctx, bucketName, objectKey, objectKey)
	if err != nil {
		return nil, err
	}
	ctx, cancelFn := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFn()
	defer os.Remove(filename)

	data, err := ffprobe.ProbeURL(ctx, filename)
	if err != nil {
		return nil, err
	}
	meta := &model.Metadata{
		Filename:       data.Format.Filename,
		NbStreams:      data.Format.NBStreams,
		NbPrograms:     data.Format.NBPrograms,
		FormatName:     data.Format.FormatName,
		FormatLongName: data.Format.FormatLongName,
		StartTime:      data.Format.StartTime().String(),
		Duration:       strconv.FormatFloat(data.Format.DurationSeconds, 'f', -1, 64),
		Size:           data.Format.Size,
		BitRate:        data.Format.BitRate,
		ProbeScore:     data.Format.ProbeScore,
		Tags: model.Tags{
			CompatibleBrands: data.Format.Tags.CompatibleBrands,
			MajorBrand:       data.Format.Tags.MajorBrand,
			MinorVersion:     data.Format.Tags.MinorVersion,
		},
	}
	log.Println(meta)
	return meta, nil
}
func (c *Controller) GetThumbnail(ctx context.Context, objectKey string) (string, error) {
	fileName, err := c.repo.DownloadPartialObject(ctx, bucketName, objectKey, objectKey, 1048575)
	if err != nil {
		return "", err
	}
	log.Printf("%s", fileName)
	defer os.Remove(fileName)
	thumbNailPath := fmt.Sprintf("%s_thumb.jpg", objectKey)

	if _, err := os.Stat(fileName); err != nil {
		log.Printf("Input file does not exist or is not accessible: %v", err)
		return "", fmt.Errorf("input file error: %w", err)
	}

	cmd := exec.Command("ffmpeg", "-ss", "00:00:01", "-t", "1", "-s", "400x300", "-i", fileName, "mjpeg", thumbNailPath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.Printf("ffmpeg error: %s", stderr.String())
		return "", fmt.Errorf("failed to generate thumbnail: %w", err)
	}
	return thumbNailPath, nil
}
func (c *Controller) GetURL(ctx context.Context, filename string) (*model.UploadURL, error) {
	objectKey := fmt.Sprintf("%s_%s", filename, time.Now().Format("20060102T150405"))
	// url, err := c.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
	// 	Bucket:      aws.String(bucketName),
	// 	Key:         aws.String(objectKey),
	// 	ContentType: aws.String("video/mp4"),
	// })

	url, err := c.repo.PutObject(ctx, bucketName, objectKey, 360)

	if err != nil {
		return nil, err
	}
	fmt.Println("Presigned URL:", url.URL)
	fmt.Println("HTTP method signed:", url.Method)
	upload := &model.UploadURL{
		JobID:        GenerateObjectKeyInt64Random(filename),
		PresignedURL: url,
		ObjectKey:    objectKey,
	}
	return upload, nil
}

// GenerateObjectKeyInt64 generates an int64 object key based on the filename.
func GenerateObjectKeyInt64(filename string) int64 {
	h := fnv.New64a()         // 64-bit FNV-1a hash
	h.Write([]byte(filename)) // hash the filename
	return int64(h.Sum64())   // convert to int64
}

func GenerateObjectKeyInt64Random(filename string) int64 {
	base := GenerateObjectKeyInt64(filename)

	var randBytes [8]byte
	_, err := rand.Read(randBytes[:])
	if err != nil {
		// fallback to deterministic base
		return base
	}
	random := int64(binary.LittleEndian.Uint64(randBytes[:]))

	// combine deterministic and random part (e.g., XOR)
	return base ^ random
}

func (c *Controller) PublishCompressionEvent(ctx context.Context, jobID int64, objectKey string, meta *model.Metadata) error {
	event := model.CompressionEvent{
		JobID:     jobID,
		ObjectKey: objectKey,
		Metadata:  *meta,
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	log.Printf("Publishing payload: %s", string(payload))
	const retries = 3
	for range retries {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		// attempt to create topic prior to publishing the message

		err := c.kafkaWriter.WriteMessages(ctx,
			kafka.Message{
				Key:   []byte(fmt.Sprintf("%d", jobID)),
				Value: payload,
			},
		)
		if errors.Is(err, kafka.LeaderNotAvailable) || errors.Is(err, context.DeadlineExceeded) {
			time.Sleep(time.Millisecond * 250)
			continue
		}
		if err != nil {
			return fmt.Errorf("kafka write error: %w", err)
		}
		break
	}

	return nil
}
