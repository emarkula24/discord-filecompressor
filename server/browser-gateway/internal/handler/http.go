package handler

import (
	"context"
	"encoding/json"
	"ffmpeg/wrapper/browser-gateway/internal/controller"
	compressionModel "ffmpeg/wrapper/compression/pkg/model"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type Handler struct {
	ctrl        *controller.VideoGatewayController
	kafkaReader *kafka.Reader
}

func NewHandler(ctrl *controller.VideoGatewayController, reader *kafka.Reader) *Handler {
	return &Handler{
		ctrl:        ctrl,
		kafkaReader: reader,
	}
}

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

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	h.kafkaReader.SetOffset(kafka.FirstOffset)
	for {
		m, err := h.kafkaReader.ReadMessage(ctx)
		if err != nil {
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
