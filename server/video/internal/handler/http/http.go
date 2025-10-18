package http

import (
	"encoding/json"
	"errors"
	"ffmpeg/wrapper/video/internal/controller/video"
	"log"
	"net/http"
)

// Handler defines a video handler.
type Handler struct {
	ctrl *video.Controller
}

// New creates a new video HTTP handler.
func New(ctrl *video.Controller) *Handler {
	return &Handler{ctrl}
}

func (h *Handler) GetCompressedVideoDetails(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Query().Get("path")
	details, err := h.ctrl.Get(req.Context(), path)
	if err != nil && errors.Is(err, video.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Repository get error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(details); err != nil {
		log.Printf("Response encode error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)

	}

}
