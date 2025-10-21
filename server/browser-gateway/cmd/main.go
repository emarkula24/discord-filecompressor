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
	"os"
	"time"

	"github.com/rs/cors"
	"github.com/segmentio/kafka-go"
)

const serviceName = "browser"

func main() {
	var port int
	flag.IntVar(&port, "port", 8085, "API handler port")
	flag.Parse()
	log.Printf("Starting the movie service on port %d", port)
	registry, err := consul.NewRegistry("consul:8500")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("gateway:%d", port)); err != nil {
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

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{os.Getenv("kafkaBroker")},
		Topic:     "compression-job",
		Partition: 0,
		MaxBytes:  10e6,
	})

	ctrl := controller.NewVideoGatewayController(gen.NewVideoServiceClient(conn))
	h := handler.NewHandler(ctrl, reader)

	mux := http.NewServeMux()

	mux.Handle("/upload", http.HandlerFunc(h.PostUploadURL))
	mux.Handle("/jobs/status", http.HandlerFunc(h.GetJobStatus))
	mux.Handle("/jobs/upload", http.HandlerFunc(h.PostUploadStatus))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://127.0.0.1:5173", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "OPTIONS"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Content-Type"},
	})
	handlerWithCORS := c.Handler(mux)

	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), handlerWithCORS); err != nil {
		panic(err)
	}
}
