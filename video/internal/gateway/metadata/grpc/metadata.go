package grpc

import (
	"context"
	"ffmpeg/wrapper/metadata/pkg/model"
	"ffmpeg/wrapper/pkg/discovery"
	"ffmpeg/wrapper/pkg/discovery/grpcutil"
	"ffmpeg/wrapper/src/gen"
)

// Gateway defines a movie metadata gRPC gateway.
type Gateway struct {
	registry discovery.Registry
}

// New creates a new gRPC gateway for a video metadata
// service.
func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry}
}

// Get returns video metadata by a movie path.
func (g *Gateway) Get(ctx context.Context, path string) (*model.Metadata, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "metadata", g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := gen.NewMetadataServiceClient(conn)
	resp, err := client.GetMetadata(ctx, &gen.GetMetadataRequest{Path: path})
	if err != nil {
		return nil, err
	}
	return model.MetadataFromProto(resp.Metadata), nil

}
