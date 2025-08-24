package middlewares

import (
	"github.com/ShlykovPavel/auth-JWT-microservice/metrics"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time"
)

func PrometheusMiddleware(metrics *metrics.Metrics) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			if r.URL.Path != "/api/v1/metrics" {
				duration := time.Since(start).Seconds()
				status := http.StatusText(ww.Status())

				metrics.HttpRequestsTotal.WithLabelValues(
					r.Method,
					r.URL.Path,
					status,
				).Inc()

				metrics.HttpRequestDuration.WithLabelValues(
					r.Method,
					r.URL.Path,
				).Observe(duration)
			}
		})
	}
}
