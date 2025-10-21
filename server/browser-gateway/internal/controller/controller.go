package controller

import (
	"context"
	"ffmpeg/wrapper/src/gen"
)

type VideoGatewayController struct {
	videoClient gen.VideoServiceClient
}

func NewVideoGatewayController(conn gen.VideoServiceClient) *VideoGatewayController {
	return &VideoGatewayController{videoClient: conn}
}

// GetUploadURL wraps gRPC call to VideoService
func (c *VideoGatewayController) GetUploadURL(ctx context.Context, filename string) (*gen.GetUploadURLResponse, error) {
	return c.videoClient.GetUploadURL(ctx, &gen.GetUploadURLRequest{Filename: filename})
}

// GetJobStatus wraps gRPC call to VideoService
func (c *VideoGatewayController) GetJobStatus(ctx context.Context, jobID int64) (*gen.GetJobStatusResponse, error) {
	return c.videoClient.GetJobStatus(ctx, &gen.GetJobStatusRequest{JobId: jobID})
}
func (c *VideoGatewayController) GetCompressionJob(ctx context.Context, jobID int64, objectKey string) (*gen.GetCompressionJobResponse, error) {
	return c.videoClient.GetCompressionJob(ctx, &gen.GetCompressionJobRequest{JobId: jobID, ObjectKey: objectKey})
}
