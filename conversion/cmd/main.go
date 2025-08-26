package main

import (
	"ffmpeg/wrapper/conversion/internal/controller/ffmpeg"
	"net/http"

	httphandler "ffmpeg/wrapper/conversion/internal/handler/http"
)

func main() {
	ctrl := ffmpeg.New()
	h := httphandler.New(ctrl)
	http.Handle("/compress", http.HandlerFunc(h.Handle))
}
