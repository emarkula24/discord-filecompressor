package handler

import (
	"context"
	"errors"
	"ffmpeg/wrapper/compression/internal/controller/ffmpeg"
	"ffmpeg/wrapper/src/gen"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	gen.UnimplementedCompressionServiceServer
	svc *ffmpeg.Controller
}

func New(ctrl *ffmpeg.Controller) *Handler {
	return &Handler{svc: ctrl}
}

func (h *Handler) GetCompressedVideoLink(ctx context.Context, req *gen.GetCompressionRequest) (*gen.GetCompressionResponse, error) {
	if req == nil || req.Videolink == "" || req.Duration == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil videolink or duration")
	}
	durationFloat, err := strconv.ParseFloat(req.Duration, 64)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	a, err := h.svc.Compress(ctx, durationFloat, req.Videolink)
	if err != nil && errors.Is(err, ffmpeg.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, "%s", err.Error())
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return &gen.GetCompressionResponse{Path: string(*a)}, nil
}
