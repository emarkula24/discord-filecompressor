package handler

import (
	"encoding/json"
	"ffmpeg/wrapper/browser-gateway/internal/controller"
	"net/http"
	"strconv"
)

type Handler struct {
	ctrl *controller.VideoGatewayController
}

func NewHandler(ctrl *controller.VideoGatewayController) *Handler {
	return &Handler{ctrl: ctrl}
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
	resp, err := h.ctrl.GetJobStatus(r.Context(), jobIDint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resp)
}
