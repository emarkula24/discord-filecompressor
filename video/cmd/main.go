package main

import (
	"context"
	"ffmpeg/wrapper/pkg/discovery"
	"ffmpeg/wrapper/pkg/discovery/consul"
	"ffmpeg/wrapper/src/gen"
	"ffmpeg/wrapper/video/internal/controller/video"
	compressiongateway "ffmpeg/wrapper/video/internal/gateway/compression/grpc"
	metadatagateway "ffmpeg/wrapper/video/internal/gateway/metadata/grpc"
	"flag"
	"fmt"
	"net"
	"time"

	grpchandler "ffmpeg/wrapper/video/internal/handler/grpc"

	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serviceName = "video"

func main() {
	var port int
	flag.IntVar(&port, "port", 8083, "API handler port")
	flag.Parse()
	log.Printf("Starting the video service on port %d", port)
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

	metadataGateway := metadatagateway.New(registry)
	compressionGateway := compressiongateway.New(registry)
	ctrl := video.New(compressionGateway, metadataGateway)

	grpcAddr := fmt.Sprintf("localhost:%d", port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	gen.RegisterVideoServiceServer(grpcServer, grpchandler.New(ctrl))
	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}
