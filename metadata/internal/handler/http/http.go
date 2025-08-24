package http

import (
	"encoding/json"
	metadata "ffmpeg/wrapper/metadata/internal/controller/meta-data"
	"log"
	"net/http"
)

type Handler struct {
	ctrl *metadata.Controller
}

func New(ctrl *metadata.Controller) *Handler {
	return &Handler{ctrl}
}

func (h *Handler) GetMetadata(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Query().Get("path")
	ctx := req.Context()
	data, err := h.ctrl.Get(ctx, path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(data.Format); err != nil {
		log.Printf("Response encode error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
