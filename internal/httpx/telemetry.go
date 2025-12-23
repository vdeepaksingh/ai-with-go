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

// TelemetryMiddleware creates HTTP middleware that records request metrics.
func TelemetryMiddleware(metrics *telemetry.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create response writer wrapper to capture status code.
			ww := &responseWriter{ResponseWriter: w, statusCode: 200}

			next.ServeHTTP(ww, r)

			duration := time.Since(start).Seconds()

			// Record HTTP request metrics.
			metrics.RecordRequest(r.Method, r.URL.Path, ww.statusCode, duration)
		})
	}
}