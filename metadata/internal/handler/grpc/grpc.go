package grpc

import (
	"context"
	"errors"
	metadata "ffmpeg/wrapper/metadata/internal/controller/metadata"
	"ffmpeg/wrapper/metadata/pkg/model"
	"ffmpeg/wrapper/src/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler defines a video metadata gRPC handler.
type Handler struct {
	gen.UnimplementedMetadataServiceServer
	gen.UnimplementedVideoServiceServer
	svc *metadata.Controller
}

// New creates a new video metadata gRPC handler.
func New(ctrl *metadata.Controller) *Handler {
	return &Handler{
		svc: ctrl,
	}
}

func (h *Handler) GetMetadata(ctx context.Context, req *gen.GetMetadataRequest) (*gen.GetMetadataResponse, error) {
	if req == nil || req.Path == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or empty path")
	}
	m, err := h.svc.GetMetadata(ctx, req.Path)
	if err != nil && errors.Is(err, metadata.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, "%s", err.Error())
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return &gen.GetMetadataResponse{Metadata: model.MetadataToProto(m)}, nil
}
func (h *Handler) GetPresignedURL(ctx context.Context, req *gen.GetUploadURLRequest) (*gen.GetUploadURLResponse, error) {
	if req == nil || req.Filename == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or empty path")
	}
	url, err := h.svc.GetURL(ctx, req.Filename)
	if err != nil && errors.Is(err, metadata.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, "%s", err.Error())
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return &gen.GetUploadURLResponse{
		JobId:        url.JobID,
		PresignedUrl: model.PresignedToProto(url.PresignedURL),
		ObjectKey:    url.ObjectKey,
	}, nil
}
