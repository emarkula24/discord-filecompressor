package main

import (
	"context"
	metadata "ffmpeg/wrapper/metadata/internal/controller/metadata"
	"ffmpeg/wrapper/pkg/discovery"
	"ffmpeg/wrapper/pkg/discovery/consul"
	"ffmpeg/wrapper/src/gen"
	"flag"
	"fmt"
	"net"
	"time"

	// httphandler "ffmpeg/wrapper/metadata/internal/handler/http"
	grpchandler "ffmpeg/wrapper/metadata/internal/handler/grpc"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/awsdocs/aws-doc-sdk-examples/gov2/s3/actions"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serviceName = "metadata"

func main() {
	var port int
	flag.IntVar(&port, "port", 8082, "API handler port")
	flag.Parse()
	log.Printf("Starting ffmpeg wrapper metadata service on port %d", port)
	registry, err := consul.NewRegistry("consul:8500")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("metadata:%d", port)); err != nil {
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

	var accountId = "ac704c26fe6c33902adaf1301c9cf006"
	var accessKeyId = "75522d5c3fb6e75fd08f54653f4dc0b3"
	var accessKey = "d45783acec00cf9ff22118b84a44c8da5a46b09ce01d9b3823cfa81c922bca42"

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId))
	})
	presignClient := s3.NewPresignClient(client)
	presigner := actions.Presigner{PresignClient: presignClient}
	ctrl := metadata.New(&presigner)
	h := grpchandler.New(ctrl)
	addr := fmt.Sprintf("metadata:%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen %v", err)
	}
	srv := grpc.NewServer()
	reflection.Register(srv)
	gen.RegisterMetadataServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
	if err != nil {
		log.Fatal(err)
	}
}
