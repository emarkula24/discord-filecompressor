package main

import (
	"context"
	"ffmpeg/wrapper/compression/internal/controller/ffmpeg"
	"ffmpeg/wrapper/compression/internal/repository"
	"ffmpeg/wrapper/gen"
	"ffmpeg/wrapper/pkg/discovery"
	"ffmpeg/wrapper/pkg/discovery/consul"
	"ffmpeg/wrapper/pkg/discovery/tracing"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	grpchandler "ffmpeg/wrapper/compression/internal/handler/grpc"
	"ffmpeg/wrapper/internal/util"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "compression"

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
	logger.Info("Starting the compression service on port %v", zap.Int("port:", port))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tp, err := tracing.NewJaegerProvider(ctx, cfg.Jaeger.URL, serviceName)
	if err != nil {
		logger.Fatal("Failed to initialize Jaeger provider", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Fatal("Failed to shut down Jaeger prodiver", zap.Error(err))
		}
	}()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	util.RecordMetrics()
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Prometheus.MetricsPort), nil); err != nil {
			logger.Fatal("Failed to start the metrics handler", zap.Error((err)))
		}
	}()

	registry, err := consul.NewRegistry(cfg.ServiceDiscovery.Consul.Address)
	if err != nil {
		panic(err)
	}
	apiInstanceID := discovery.GenerateInstanceID(serviceName + "-api")
	metricsInstanceID := discovery.GenerateInstanceID(serviceName + "-metrics")

	err = registry.Register(ctx, metricsInstanceID, serviceName+"-metrics", fmt.Sprintf("%s:%d", serviceName, cfg.Prometheus.MetricsPort))
	if err != nil {
		panic(err)
	}

	err = registry.Register(ctx, apiInstanceID, serviceName+"-api", fmt.Sprintf("%s:%d", serviceName, port))
	if err != nil {
		panic(err)
	}
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := registry.ReportHealthyState(apiInstanceID, "api"); err != nil {
				logger.Error("Failed to report healthy state", zap.Error(err))
			}
			if err := registry.ReportHealthyState(metricsInstanceID, "metrics"); err != nil {
				logger.Error("Failed to report healthy state", zap.Error(err))
			}
		}

	}()
	defer func() {
		if err := registry.Deregister(ctx, apiInstanceID, serviceName+"-api"); err != nil {
			logger.Error("Failed to deregister API service", zap.Error(err))
		}
		if err := registry.Deregister(ctx, metricsInstanceID, serviceName+"-metrics"); err != nil {
			logger.Error("Failed to deregister Metrics service", zap.Error(err))
		}
	}()
	var accountId = os.Getenv("accountId")
	var accessKeyId = os.Getenv("accessKeyId")
	var accessKey = os.Getenv("secretKey")

	AWScfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		logger.Error("Failed to load AWS s3 config", zap.Error(err))
	}

	s3Client := s3.NewFromConfig(AWScfg, func(o *s3.Options) {
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
		logger.Fatal("Failed to listen", zap.Error(err))
	}
	srv := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
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
