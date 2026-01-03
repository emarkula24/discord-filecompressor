package grpc

import (
	"context"
	"errors"
	"ffmpeg/wrapper/gen"
	metadata "ffmpeg/wrapper/metadata/internal/controller/metadata"
	"ffmpeg/wrapper/metadata/pkg/model"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler defines a video metadata gRPC handler.
type Handler struct {
	gen.UnimplementedMetadataServiceServer
	svc *metadata.Controller
}

// New creates a new video metadata gRPC handler.
func New(ctrl *metadata.Controller) *Handler {
	return &Handler{
		svc: ctrl,
	}
}
func (h *Handler) GetCompressionJob(ctx context.Context, req *gen.GetCompressionJobRequest) (*gen.GetCompressionJobResponse, error) {
	if req == nil || req.ObjectKey == "" {
		return nil, status.Error(codes.InvalidArgument, "nil req or empty objectkey")
	}
	m, err := h.svc.GetMetadata(ctx, req.ObjectKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	fmt.Println(m)
	err = h.svc.PublishCompressionEvent(ctx, req.JobId, req.ObjectKey, m)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return &gen.GetCompressionJobResponse{
		JobId:  req.JobId,
		Status: "started",
	}, nil
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
func (h *Handler) GetUploadURL(ctx context.Context, req *gen.GetUploadURLRequest) (*gen.GetUploadURLResponse, error) {
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
