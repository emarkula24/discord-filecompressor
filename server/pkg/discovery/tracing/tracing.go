package tracing

import (
	"context"

	jaeger "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"

	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

// NewJaegerProvider returns a new jaeger-based tracing provider.
func NewJaegerProvider(ctx context.Context, endpoint string, serviceName string) (*tracesdk.TracerProvider, error) {
	exp, err := jaeger.New(ctx,
		jaeger.WithEndpoint(endpoint),
		jaeger.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)
	return tp, nil
}
