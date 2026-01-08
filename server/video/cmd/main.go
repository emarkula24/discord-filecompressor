package main

import (
	"context"
	"ffmpeg/wrapper/gen"
	"ffmpeg/wrapper/pkg/discovery"
	"ffmpeg/wrapper/pkg/discovery/consul"
	"ffmpeg/wrapper/pkg/discovery/tracing"
	"ffmpeg/wrapper/video/internal/controller/video"
	compressiongateway "ffmpeg/wrapper/video/internal/gateway/compression/grpc"
	metadatagateway "ffmpeg/wrapper/video/internal/gateway/metadata/grpc"
	"fmt"
	"net"
	"os"
	"time"

	grpchandler "ffmpeg/wrapper/video/internal/handler/grpc"

	"log"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "video"

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	f, err := os.Open("../configs/default.yaml")
	if err != nil {
		logger.Fatal("Failed to open configuration file", zap.Error(err))
	}
	var cfg configuration
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		logger.Fatal("Failed to decode configuration file", zap.Error(err))
	}
	port := cfg.API.Port
	logger.Info("Starting the video service", zap.Int("port", port))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tp, err := tracing.NewJaegerProvider(ctx, cfg.Jaeger.URL, serviceName)
	if err != nil {
		logger.Fatal("Failed to initialize Jaeger provider", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal("Failed to shut down Jaeger prodiver", zap.Error(err))
		}
	}()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	registry, err := consul.NewRegistry(cfg.ServiceDiscovery.Consul.Address)
	if err != nil {
		panic(err)
	}

	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("video:%d", port)); err != nil {
		panic(err)
	}
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				logger.Error("Failed to report healthy state", zap.Error(err))
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)

	metadataGateway := metadatagateway.New(registry)
	compressionGateway := compressiongateway.New(registry)
	ctrl := video.New(compressionGateway, metadataGateway)

	grpcAddr := fmt.Sprintf("0.0.0.0:%d", port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal("Failed to listen", zap.Error(err))
	}
	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	reflection.Register(grpcServer)
	gen.RegisterVideoServiceServer(grpcServer, grpchandler.New(ctrl))
	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}
