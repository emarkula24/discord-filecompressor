package grpc

import (
	"context"
	"errors"
	"ffmpeg/wrapper/metadata/pkg/model"
	"ffmpeg/wrapper/src/gen"
	"ffmpeg/wrapper/video/internal/controller/video"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler defines a video gRPC handler.
type Handler struct {
	gen.UnimplementedVideoServiceServer
	svc *video.Controller
}

// New creates a new video gRPC handler.
func New(ctrl *video.Controller) *Handler {
	return &Handler{svc: ctrl}
}

func (h *Handler) GetCompressedVideoDetails(ctx context.Context, req *gen.GetVideoDetailsRequest) (*gen.GetVideoDetailsResponse, error) {
	if req == nil || req.Path == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or empty path")
	}
	m, err := h.svc.Get(ctx, req.Path)
	if err != nil && errors.Is(err, video.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, "%s", err.Error())
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return &gen.GetVideoDetailsResponse{
			Link:        m.Link,
			OldMetadata: model.MetadataToProto(&m.OldMetadata),
		},
		nil
}
func (h *Handler) GetUploadURL(ctx context.Context, req *gen.GetUploadURLRequest) (*gen.GetUploadURLResponse, error) {
	return h.svc.GetUploadURL(ctx, req)
}
