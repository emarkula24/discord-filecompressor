package main

import (
	"context"
	"ffmpeg/wrapper/compression/internal/controller/ffmpeg"
	"ffmpeg/wrapper/pkg/discovery"
	"ffmpeg/wrapper/pkg/discovery/consul"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	httphandler "ffmpeg/wrapper/compression/internal/handler/http"
)

const serviceName = "compression"

func main() {
	var port int
	flag.IntVar(&port, "port", 8081, "API handler port")
	flag.Parse()
	log.Printf("Starting the compression service on port %d", port)
	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
		panic(err)
	}
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Println("Failed to report healthy state: " + err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)
	ctrl := ffmpeg.New()
	h := httphandler.New(ctrl)
	http.Handle("/compress", http.HandlerFunc(h.Handle))
	if err := http.ListenAndServe(fmt.Sprintf("%d", port), nil); err != nil {
		panic(err)
	}

}
