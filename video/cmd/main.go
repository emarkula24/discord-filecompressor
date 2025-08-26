package cmd

import (
	"ffmpeg/wrapper/video/internal/controller/video"
	compressiongateway "ffmpeg/wrapper/video/internal/gateway/conversion/http"
	metadatagateway "ffmpeg/wrapper/video/internal/gateway/metadata/http"
	httphandler "ffmpeg/wrapper/video/internal/handler/http"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting the video service")
	metadataGateway := metadatagateway.New("localhost:8081")
	compressionGateway := compressiongateway.New("localhost:8082")
	ctrl := video.New(compressionGateway, metadataGateway)
	h := httphandler.New(ctrl)
	http.Handle("/video", http.HandlerFunc(h.GetCompressedVideoDetails))
	if err := http.ListenAndServe(":8083", nil); err != nil {
		panic(err)
	}

}
