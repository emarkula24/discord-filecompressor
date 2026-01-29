package util

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// func NewResource() (*resource.Resource, error) {
// 	return resource.Merge(
// 		resource.Default(),
// 		resource.NewWithAttributes(
// 			semconv.SchemaURL,
// 			semconv.ServiceName("my-service"),
// 			semconv.ServiceVersion("0.1.0"),
// 		),
// 	)
// }

// func NewMeterProvider(res *resource.Resource) (*metric.MeterProvider, error) {
// 	metricExporter, err := stdoutmetric.New()
// 	if err != nil {
// 		return nil, err
// 	}

// 	meterProvider := metric.NewMeterProvider(
// 		metric.WithResource(res),
// 		metric.WithReader(metric.NewPeriodicReader(metricExporter,
// 			// Default is 1m. Set to 3s for demonstrative purposes.
// 			metric.WithInterval(3*time.Second))),
// 	)
// 	return meterProvider, nil
// }

var once sync.Once

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "processed_operations_total",
		Help: "The total number of processed events",
	})
)

func RecordMetrics() {
	// Ensure we only start the recording loop once
	once.Do(func() {
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				opsProcessed.Inc()
			}
		}()
	})
}

// IncrementOpsProcessed increments the operations counter
func IncrementOpsProcessed() {
	opsProcessed.Inc()
}
