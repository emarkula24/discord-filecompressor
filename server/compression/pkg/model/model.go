package model

type Duration float64
type VideoLink string
type CompressedLink string

type CompressionResultEvent struct {
	CompressionEventType CompressionEventType     `json:"event_type"`
	JobID                int64                    `json:"job_id"`
	CompressedKey        string                   `json:"compressed_key"`
	ObjectKey            string                   `json:"object_key"`
	PresignedDownloadUrl *PresignedRequestPayload `json:"presigned_download_url"`
}

type CompressionEventType string

const (
	CompressionEventTypeSuccess = CompressionEventType("success")
	CompressionEventTypeFail    = CompressionEventType("fail")
)

type PresignedRequestPayload struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
}
