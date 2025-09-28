package metadata

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"ffmpeg/wrapper/metadata/pkg/model"
	"fmt"
	"hash/fnv"
	"path/filepath"
	"time"

	"github.com/awsdocs/aws-doc-sdk-examples/gov2/s3/actions"
	"gopkg.in/vansante/go-ffprobe.v2"
)

var bucketName = "compressor-g"
var ErrNotFound = errors.New("not found")

type Controller struct {
	presigner *actions.Presigner
}

func New(presn *actions.Presigner) *Controller {
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
	objectkey := GenerateObjectKeyString(filename)
	url, err := c.presigner.PutObject(ctx, bucketName, filename, 60)
	if err != nil {
		return nil, err
	}
	upload := &model.UploadURL{
		JobID:        GenerateObjectKeyInt64Random(filename),
		PresignedURL: url,
		ObjectKey:    objectkey,
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

func GenerateObjectKeyString(filename string) string {
	// Create a numeric key
	keyNum := GenerateObjectKeyInt64Random(filename)

	// Use the base filename + numeric key as the object key
	base := filepath.Base(filename)
	return fmt.Sprintf("%s_%d", base, keyNum)
}
