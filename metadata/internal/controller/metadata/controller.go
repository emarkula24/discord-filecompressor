package metadata

import (
	"context"
	"errors"
	"ffmpeg/wrapper/metadata/pkg/model"
	"time"

	"gopkg.in/vansante/go-ffprobe.v2"
)

var ErrNotFound = errors.New("not found")

type Controller struct{}

func New() *Controller {
	return &Controller{}
}
func (c *Controller) Get(ctx context.Context, path string) (*model.Metadata, error) {
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
