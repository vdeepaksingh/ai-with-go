package httpx

import (
	"net/http"
	"time"

	"github.com/acai-travel/tech-challenge/internal/telemetry"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func TelemetryMiddleware(metrics *telemetry.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			ww := &responseWriter{ResponseWriter: w, statusCode: 200}

			next.ServeHTTP(ww, r)

			duration := time.Since(start).Seconds()

			// Record metrics
			metrics.RecordRequest(r.Method, r.URL.Path, ww.statusCode, duration)
		})
	}
}