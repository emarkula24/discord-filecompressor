package handler

import (
	"context"
	"encoding/json"
	"errors"
	compressionModel "ffmpeg/wrapper/compression/pkg/model"
	"ffmpeg/wrapper/gateway/internal/controller"
	"ffmpeg/wrapper/gateway/internal/repository"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/segmentio/kafka-go"
)

type Handler struct {
	ctrl        *controller.VideoGatewayController
	kafkaReader *kafka.Reader
	repo        repository.S3Actions
}

func NewHandler(ctrl *controller.VideoGatewayController, reader *kafka.Reader, repo repository.S3Actions) *Handler {
	return &Handler{
		ctrl:        ctrl,
		kafkaReader: reader,
		repo:        repo,
	}
}

var bucketName = os.Getenv("bucketname")

// POST /upload
func (h *Handler) PostUploadURL(w http.ResponseWriter, r *http.Request) {
	var req struct{ Filename string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	resp, err := h.ctrl.GetUploadURL(r.Context(), req.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resp)
}

// GET /jobs/status?job_id=xxx
func (h *Handler) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("job_id")
	if jobID == "" {
		http.Error(w, "missing job_id", http.StatusBadRequest)
		return
	}
	jobIDint, err := strconv.ParseInt(jobID, 10, 64)
	if err != nil {
		http.Error(w, "failed convert", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	h.kafkaReader.SetOffset(kafka.FirstOffset)
	for {
		m, err := h.kafkaReader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				json.NewEncoder(w).Encode(map[string]string{
					"status": "processing",
				})
				return
			}

			log.Printf("kafka reading error: %v", err)
			http.Error(w, "error reading Kafka message", http.StatusInternalServerError)
			return
		}
		var result compressionModel.CompressionResultEvent
		if err := json.Unmarshal(m.Value, &result); err != nil {
			continue // skip malformed messages
		}
		if result.JobID == jobIDint && result.CompressionEventType != "" {
			json.NewEncoder(w).Encode((result))

			go func(expiry time.Time, objectKeys []string) {
				log.Print(expiry, objectKeys)
				duration := time.Until(expiry)
				log.Print(duration)
				if duration > 0 {
					time.Sleep(duration)
				}
				objs := make([]types.ObjectIdentifier, 0, len(objectKeys))
				for _, k := range objectKeys {
					objs = append(objs, types.ObjectIdentifier{Key: aws.String(k)})
				}
				bgCtx := context.Background()
				h.repo.DeleteObjects(bgCtx, bucketName, objs, false)
			}(result.Expiry, []string{result.ObjectKey, result.CompressedKey})

			return
		}
	}
}

func (h *Handler) PostUploadStatus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		JobID     int64  `json:"job_id"`
		ObjectKey string `json:"object_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
	}
	resp, err := h.ctrl.GetCompressionJob(r.Context(), req.JobID, req.ObjectKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(resp)
	json.NewEncoder(w).Encode(resp)
}
