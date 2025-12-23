package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
)

// InitTracing initializes OpenTelemetry tracing with stdout exporter.
func InitTracing() {
	// Create stdout exporter for development tracing.
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		slog.Error("Failed to create trace exporter", "error", err)
		return
	}

	// Create trace provider with batch exporter.
	provider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(provider)

	slog.Info("Tracing initialized")
}

// Shutdown gracefully shuts down the OpenTelemetry trace provider.
func Shutdown(ctx context.Context) {
	if tp, ok := otel.GetTracerProvider().(*trace.TracerProvider); ok {
		tp.Shutdown(ctx)
	}
}