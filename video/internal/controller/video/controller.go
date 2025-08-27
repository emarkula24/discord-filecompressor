package video

import (
	"context"
	"errors"
	conversionmodel "ffmpeg/wrapper/compression/pkg/model"
	metadatamodel "ffmpeg/wrapper/metadata/pkg/model"
	"ffmpeg/wrapper/video/internal/gateway"
	"ffmpeg/wrapper/video/pkg/model"
)

// ErrNotFound is returned when the video metadata is not
// found.
var ErrNotFound = errors.New("movie metadata not found")

type compressionGateway interface {
	GetCompressedVideo(ctx context.Context, duration conversionmodel.Duration, videoLink conversionmodel.VideoLink) (string, error)
}
type metadataGateway interface {
	Get(ctx context.Context, path string) (*metadatamodel.Metadata, error)
}

// Controller defines a video service controller.
type Controller struct {
	compressionGateway compressionGateway
	metadataGateway    metadataGateway
}

// New creates a new videservice controller.
func New(compressionGateway compressionGateway, metadataGateway metadataGateway) *Controller {
	return &Controller{compressionGateway, metadataGateway}

}

// Get returns the movie details including the aggregated
// rating and movie metadata.
// Get returns the movie details including the aggregated rating and movie metadat
func (c *Controller) Get(ctx context.Context, path string) (*model.ConvertedVideo, error) {

	metadata, err := c.metadataGateway.Get(ctx, path)
	if err != nil && errors.Is(err, gateway.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	details := &model.ConvertedVideo{OldMetadata: *metadata}

	convertedVideoLink, err := c.compressionGateway.GetCompressedVideo(ctx, conversionmodel.Duration(metadata.Duration), conversionmodel.VideoLink(path))

	if err != nil && !errors.Is(err, gateway.ErrNotFound) {
		// Just proceed in this case, it's ok not to have videos yet.
	} else if err != nil {
		return nil, err
	} else {

		details.Link = convertedVideoLink

	}

	return details, nil

}
