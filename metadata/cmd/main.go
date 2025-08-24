package main

import (
	metadata "ffmpeg/wrapper/metadata/internal/controller/meta-data"
	httphandler "ffmpeg/wrapper/metadata/internal/handler/http"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting ffmpeg wrapper metadata service")
	ctrl := metadata.New()
	h := httphandler.New(ctrl)
	http.Handle("/metadata", http.HandlerFunc(h.GetMetadata))
	if err := http.ListenAndServe(":8081", nil); err != nil {

		panic(err)

	}

}

// d, err := ctrl.Get(ctx, "/mnt/F6ECB1FBECB1B669/Videot/shadman.webm")
