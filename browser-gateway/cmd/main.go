package main

import (
	"context"
	"ffmpeg/wrapper/browser-gateway/internal/controller"
	"ffmpeg/wrapper/browser-gateway/internal/handler"
	"ffmpeg/wrapper/pkg/discovery"
	"ffmpeg/wrapper/pkg/discovery/consul"
	"ffmpeg/wrapper/pkg/discovery/grpcutil"
	"ffmpeg/wrapper/src/gen"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

const serviceName = "browser"

func main() {
	var port int
	flag.IntVar(&port, "port", 8085, "API handler port")
	flag.Parse()
	log.Printf("Starting the movie service on port %d", port)
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
	conn, err := grpcutil.ServiceConnection(ctx, "video", registry)
	if err != nil {
		log.Fatalf("failed to connect to VideoService: %v", err)
	}
	defer conn.Close()

	ctrl := controller.NewVideoGatewayController(gen.NewVideoServiceClient(conn))
	h := handler.NewHandler(ctrl)
	http.Handle("/upload", http.HandlerFunc(h.PostUploadURL))
	http.Handle("/jobs/status", http.HandlerFunc(h.GetJobStatus))

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}
