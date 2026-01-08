package main

import (
	"context"
	"ffmpeg/wrapper/gateway/internal/controller"
	"ffmpeg/wrapper/gateway/internal/handler"
	"ffmpeg/wrapper/gateway/internal/repository"
	"ffmpeg/wrapper/gen"
	"ffmpeg/wrapper/internal/grpcutil"
	"ffmpeg/wrapper/pkg/discovery"
	"ffmpeg/wrapper/pkg/discovery/consul"
	"ffmpeg/wrapper/pkg/discovery/tracing"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/cors"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const serviceName = "browser"

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	f, err := os.Open("../configs/default.yaml")
	if err != nil {
		logger.Fatal("Failed to open configuration", zap.Error(err))
	}
	var cfg configuration
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		logger.Fatal("Failed to parse configuration", zap.Error(err))
	}
	port := cfg.API.Port
	logger.Info("Starting the metadata service", zap.Int("port", port))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tp, err := tracing.NewJaegerProvider(ctx, cfg.Jaeger.URL, serviceName)
	if err != nil {
		logger.Fatal("Failed to initialize Jaeger provider", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Fatal("Failed to shutdown Jaeger provider", zap.Error(err))
		}
	}()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	instanceID := discovery.GenerateInstanceID(serviceName)

	registry, err := consul.NewRegistry(cfg.ServiceDiscovery.Consul.Address)
	if err != nil {
		panic(err)
	}

	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("gateway:%d", port)); err != nil {
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
	conn, err := grpcutil.ServiceConnection(ctx, "video", registry)
	if err != nil {
		logger.Fatal("Failed to connect to VideoService", zap.Error(err))
	}
	defer conn.Close()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{os.Getenv("kafkaBroker")},
		Topic:     "compression-job",
		Partition: 0,
		MaxBytes:  10e6,
	})
	var accountId = os.Getenv("accountId")
	var accessKeyId = os.Getenv("accessKeyId")
	var accessKey = os.Getenv("secretKey")

	AWScfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		log.Fatal(err)
	}
	s3Client := s3.NewFromConfig(AWScfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId))
		o.UsePathStyle = true
	})

	repo := repository.New(s3Client)
	ctrl := controller.NewVideoGatewayController(gen.NewVideoServiceClient(conn))
	h := handler.NewHandler(ctrl, reader, repo)

	mux := http.NewServeMux()

	// handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	// 	handler := otelhttp.WithMetricAttributesFn(http.HandlerFunc(handlerFunc))
	// 	mux.Handle(pattern, handler)
	// }

	// handleFunc("/upload", h.PostUploadURL)
	// handleFunc("/jobs/status", h.GetJobStatus)
	// handleFunc("/jobs/upload", h.PostUploadStatus)

	mux.Handle("/upload", http.HandlerFunc(h.PostUploadURL))
	mux.Handle("/jobs/status", http.HandlerFunc(h.GetJobStatus))
	mux.Handle("/jobs/upload", http.HandlerFunc(h.PostUploadStatus))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://127.0.0.1:5173", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "OPTIONS"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"*"},
	})

	// handlerWithInstruments := otelhttp.NewHandler(mux, "/")
	// handlerWithCORS := c.Handler(handlerWithInstruments)
	handlerWithCORS := c.Handler(mux)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), handlerWithCORS); err != nil {
		logger.Fatal("Failed to startup listen and serve", zap.Error(err))
	}
}
