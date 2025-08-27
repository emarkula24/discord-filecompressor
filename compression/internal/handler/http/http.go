package http

import (
	"ffmpeg/wrapper/compression/internal/controller/ffmpeg"
	"net/http"
	"strconv"
)

type Handler struct {
	ctrl *ffmpeg.Controller
}

func New(ctrl *ffmpeg.Controller) *Handler {
	return &Handler{ctrl}
}

func (h *Handler) Handle(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		videoLink := req.URL.Query().Get("videolink")
		duration := req.URL.Query().Get("duration")
		durationFloat, err := strconv.ParseFloat(duration, 64)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		err = h.ctrl.Compress(req.Context(), durationFloat, videoLink)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
