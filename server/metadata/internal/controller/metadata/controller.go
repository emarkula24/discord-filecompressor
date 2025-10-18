package metadata

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"ffmpeg/wrapper/metadata/pkg/model"
	"fmt"
	"hash/fnv"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gopkg.in/vansante/go-ffprobe.v2"
)

var bucketName = os.Getenv("bucketname")
var ErrNotFound = errors.New("not found")

type Controller struct {
	presigner *s3.PresignClient
}

func New(presn *s3.PresignClient) *Controller {
	return &Controller{
		presigner: presn,
	}
}
func (c *Controller) GetMetadata(ctx context.Context, path string) (*model.Metadata, error) {
	ctx, cancelFn := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFn()

	data, err := ffprobe.ProbeURL(ctx, path)
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
		Duration:       data.Format.StartTime().String(),
		Size:           data.Format.Size,
		BitRate:        data.Format.BitRate,
		ProbeScore:     data.Format.ProbeScore,
		Tags: model.Tags{
			CompatibleBrands: data.Format.Tags.CompatibleBrands,
			MajorBrand:       data.Format.Tags.MajorBrand,
			MinorVersion:     data.Format.Tags.MinorVersion,
		},
	}

	return meta, nil
}
func (c *Controller) GetURL(ctx context.Context, filename string) (*model.UploadURL, error) {
	objectKey := fmt.Sprintf("%s_%s", filename, time.Now().Format("20060102T150405"))
	url, err := c.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		ContentType: aws.String("video/mp4"),
	})

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
