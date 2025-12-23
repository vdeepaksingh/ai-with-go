// Package telemetry provides OpenTelemetry metrics and tracing functionality.
package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// Metrics holds OpenTelemetry metric instruments for HTTP monitoring.
type Metrics struct {
	requestCount    metric.Int64Counter
	requestDuration metric.Float64Histogram
	errorCount      metric.Int64Counter
}

// NewMetrics creates and configures OpenTelemetry metrics with Prometheus exporter.
func NewMetrics() *Metrics {
	// Create Prometheus exporter for metrics collection.
	exporter, err := prometheus.New()
	if err != nil {
		slog.Error("Failed to create Prometheus exporter", "error", err)
		return nil
	}

	// Create meter provider with Prometheus reader.
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	otel.SetMeterProvider(provider)

	// Create meter for chat service metrics.
	meter := otel.Meter("chat-service")

	// Create core HTTP metrics counters and histograms.
	requestCount, _ := meter.Int64Counter("http_requests_total")
	requestDuration, _ := meter.Float64Histogram("http_request_duration_seconds")
	errorCount, _ := meter.Int64Counter("http_errors_total")

	return &Metrics{
		requestCount:    requestCount,
		requestDuration: requestDuration,
		errorCount:      errorCount,
	}
}

// RecordRequest records HTTP request metrics including count, duration, and errors.
func (m *Metrics) RecordRequest(method, path string, statusCode int, duration float64) {
	if m == nil {
		return
	}

	ctx := context.Background()
	attrs := []attribute.KeyValue{
		attribute.String("method", method),
		attribute.String("path", path),
		attribute.String("status", fmt.Sprintf("%d", statusCode)),
	}

	// Record request count.
	m.requestCount.Add(ctx, 1, metric.WithAttributes(attrs...))

	// Record request duration.
	m.requestDuration.Record(ctx, duration, metric.WithAttributes(attrs...))

	// Record errors (4xx and 5xx status codes).
	if statusCode >= 400 {
		m.errorCount.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}