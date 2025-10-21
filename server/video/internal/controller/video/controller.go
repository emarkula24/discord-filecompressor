package video

import (
	"context"
	"errors"
	conversionmodel "ffmpeg/wrapper/compression/pkg/model"
	metadatamodel "ffmpeg/wrapper/metadata/pkg/model"
	"ffmpeg/wrapper/src/gen"
	"ffmpeg/wrapper/video/internal/gateway"
	"ffmpeg/wrapper/video/pkg/model"
	"log"
	"strconv"
)

// ErrNotFound is returned when the video metadata is not
// found.
var ErrNotFound = errors.New("movie metadata not found")

type compressionGateway interface {
	GetCompressedVideo(ctx context.Context, duration conversionmodel.Duration, videoLink conversionmodel.VideoLink) (string, error)
}
type metadataGateway interface {
	Get(ctx context.Context, path string) (*metadatamodel.Metadata, error)
	GetPresignedURL(ctx context.Context, r *gen.GetUploadURLRequest) (*gen.GetUploadURLResponse, error)
	GetCompressionJob(ctx context.Context, r *gen.GetCompressionJobRequest) (*gen.GetCompressionJobResponse, error)
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

// Get returns the movie details including the aggregated rating and movie metadata
func (c *Controller) Get(ctx context.Context, path string) (*model.ConvertedVideo, error) {

	metadata, err := c.metadataGateway.Get(ctx, path)
	if err != nil && errors.Is(err, gateway.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	details := &model.ConvertedVideo{OldMetadata: *metadata}
	durationFloat, err := strconv.ParseFloat(metadata.Duration, 64)
	if err != nil {
		log.Printf("failed to parse duration: %v", err)
		return nil, err
	}
	convertedVideoLink, err := c.compressionGateway.GetCompressedVideo(ctx, conversionmodel.Duration(durationFloat), conversionmodel.VideoLink(path))

	if err != nil && !errors.Is(err, gateway.ErrNotFound) {
		// Just proceed in this case, it's ok not to have videos yet.
	} else if err != nil {
		return nil, err
	} else {

		details.Link = convertedVideoLink

	}

	return details, nil

}
func (c *Controller) GetUploadURL(ctx context.Context, req *gen.GetUploadURLRequest) (*gen.GetUploadURLResponse, error) {
	return c.metadataGateway.GetPresignedURL(ctx, req)
}

func (c *Controller) GetCompressionJob(ctx context.Context, req *gen.GetCompressionJobRequest) (*gen.GetCompressionJobResponse, error) {
	return c.metadataGateway.GetCompressionJob(ctx, req)
}
