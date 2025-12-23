package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/acai-travel/tech-challenge/internal/chat"
	"github.com/acai-travel/tech-challenge/internal/chat/assistant"
	"github.com/acai-travel/tech-challenge/internal/chat/model"
	"github.com/acai-travel/tech-challenge/internal/httpx"
	"github.com/acai-travel/tech-challenge/internal/mongox"
	"github.com/acai-travel/tech-challenge/internal/pb"
	"github.com/acai-travel/tech-challenge/internal/telemetry"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/twitchtv/twirp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	// Initialize telemetry (metrics + tracing)
	telemetry.InitTracing()
	metrics := telemetry.NewMetrics()

	mongo := mongox.MustConnect()
	repo := model.New(mongo)
	assist := assistant.New()
	server := chat.NewServer(repo, assist)

	// Configure handler with telemetry
	handler := mux.NewRouter()
	handler.Use(
		httpx.Logger(),
		httpx.Recovery(),
		httpx.TelemetryMiddleware(metrics),
	)

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "Hi, my name is Clippy!")
	})

	// Metrics endpoint
	handler.Handle("/metrics", promhttp.Handler())

	// Twirp handler with automatic tracing
	twirpHandler := pb.NewChatServiceServer(server, twirp.WithServerJSONSkipDefaults(true))
	handler.PathPrefix("/twirp/").Handler(otelhttp.NewHandler(twirpHandler, "chat-api"))

	// Start server
	slog.Info("Starting server with metrics and tracing...")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		panic(err)
	}
}
