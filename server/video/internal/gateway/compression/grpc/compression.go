package grpc

import (
	"context"
	"ffmpeg/wrapper/compression/pkg/model"
	"ffmpeg/wrapper/gen"
	"ffmpeg/wrapper/internal/grpcutil"
	"ffmpeg/wrapper/pkg/discovery"
	"strconv"
)

// Gateway defines an gRPC gateway for a rating service.
type Gateway struct {
	registry discovery.Registry
}

// New creates a new gRPC gateway for a rating service.
func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry}
}

// GetAggregatedRating returns the aggregated rating for a
// record or ErrNotFound if there are no ratings for it.
func (g *Gateway) GetCompressedVideo(ctx context.Context, duration model.Duration, videoLink model.VideoLink) (string, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "compression", g.registry)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	client := gen.NewCompressionServiceClient(conn)
	resp, err := client.GetCompression(ctx, &gen.GetCompressionRequest{Videolink: string(videoLink), Duration: strconv.FormatFloat(float64(duration), 'f', -1, 64)})
	if err != nil {
		return "", err
	}
	return resp.Path, nil
}
