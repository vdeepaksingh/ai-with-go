package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
)

func InitTracing() {
	// Simple stdout exporter for tracing (can be replaced with Jaeger/etc)
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		slog.Error("Failed to create trace exporter", "error", err)
		return
	}

	// Simple trace provider
	provider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(provider)

	slog.Info("Tracing initialized")
}

func Shutdown(ctx context.Context) {
	if tp, ok := otel.GetTracerProvider().(*trace.TracerProvider); ok {
		tp.Shutdown(ctx)
	}
}