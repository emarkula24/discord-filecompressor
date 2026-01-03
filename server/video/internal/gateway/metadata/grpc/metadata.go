package grpc

import (
	"context"
	"ffmpeg/wrapper/gen"
	"ffmpeg/wrapper/internal/grpcutil"
	"ffmpeg/wrapper/metadata/pkg/model"
	"ffmpeg/wrapper/pkg/discovery"
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

func (g *Gateway) GetPresignedURL(ctx context.Context, req *gen.GetUploadURLRequest) (*gen.GetUploadURLResponse, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "metadata", g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := gen.NewMetadataServiceClient(conn)
	resp, err := client.GetUploadURL(ctx, &gen.GetUploadURLRequest{Filename: req.Filename})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (g *Gateway) GetCompressionJob(ctx context.Context, req *gen.GetCompressionJobRequest) (*gen.GetCompressionJobResponse, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "metadata", g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := gen.NewMetadataServiceClient(conn)
	resp, err := client.GetCompressionJob(ctx, &gen.GetCompressionJobRequest{JobId: req.JobId, ObjectKey: req.ObjectKey})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
