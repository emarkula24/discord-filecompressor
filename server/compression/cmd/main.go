package main

import (
	"context"
	"ffmpeg/wrapper/compression/internal/controller/ffmpeg"
	"ffmpeg/wrapper/pkg/discovery"
	"ffmpeg/wrapper/pkg/discovery/consul"
	"ffmpeg/wrapper/src/gen"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	grpchandler "ffmpeg/wrapper/compression/internal/handler/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serviceName = "compression"

func main() {
	var port int
	flag.IntVar(&port, "port", 8081, "API handler port")
	flag.Parse()
	log.Printf("Starting the compression service on port %d", port)
	registry, err := consul.NewRegistry("consul:8500")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("compression:%d", port)); err != nil {
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
	h := grpchandler.New(ctrl)
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	reflection.Register(srv)
	gen.RegisterCompressionServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
