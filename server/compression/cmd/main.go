package main

import (
	"context"
	"ffmpeg/wrapper/compression/internal/controller/ffmpeg"
	"ffmpeg/wrapper/compression/internal/repository"
	"ffmpeg/wrapper/gen"
	"ffmpeg/wrapper/pkg/discovery"
	"ffmpeg/wrapper/pkg/discovery/consul"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	grpchandler "ffmpeg/wrapper/compression/internal/handler/grpc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/segmentio/kafka-go"
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

	var accountId = os.Getenv("accountId")
	var accessKeyId = os.Getenv("accessKeyId")
	var accessKey = os.Getenv("secretKey")

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		log.Fatal(err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId))
		o.UsePathStyle = true
	})
	presignClient := s3.NewPresignClient(s3Client)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{os.Getenv("kafkaBroker")},
		Topic:       "compression-job",
		GroupID:     "compression-worker",
		MaxBytes:    10e6,
		StartOffset: kafka.LastOffset,
	})
	writer := &kafka.Writer{
		Addr:        kafka.TCP(os.Getenv("kafkaBroker")),
		Topic:       "compression-job",
		Balancer:    &kafka.LeastBytes{},
		Logger:      kafka.LoggerFunc(logf),
		ErrorLogger: kafka.LoggerFunc(logf),
	}
	repo := repository.New(presignClient, s3Client)
	ctrl := ffmpeg.New(reader, writer, repo)
	go ctrl.ConsumeCompressionEvent(ctx)
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

func logf(msg string, a ...interface{}) {
	fmt.Printf(msg, a...)
	fmt.Println()
}
