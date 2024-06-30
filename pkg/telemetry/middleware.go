package telemetry

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w, http.StatusOK}
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()
		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)
		duration := time.Since(start)

		logrus.WithFields(logrus.Fields{
			"method":   r.Method,
			"url":      r.URL.Path,
			"duration": duration,
			"status":   rw.statusCode,
		}).Info("Handled request")
	})
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		meter := otel.Meter("http-metrics")
		requestCount, _ := meter.Int64Counter("http.requests.total", metric.WithDescription("Total number of HTTP requests"), metric.WithUnit("{request}"))
		requestDuration, _ := meter.Float64Histogram("http.request.duration.seconds", metric.WithDescription("HTTP request duration in seconds"), metric.WithUnit("s"))

		start := time.Now()
		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)
		duration := time.Since(start).Seconds()

		attrs := []attribute.KeyValue{
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.Path),
			attribute.Int("http.status_code", rw.statusCode),
		}

		requestCount.Add(r.Context(), 1, metric.WithAttributes(attrs...))
		requestDuration.Record(r.Context(), duration, metric.WithAttributes(attrs...))
	})
}
