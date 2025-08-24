package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics структура с метриками
type Metrics struct {
	HttpRequestsTotal   *prometheus.CounterVec
	HttpRequestDuration *prometheus.HistogramVec
	PgxPoolMaxConns     prometheus.Gauge
	PgxPoolUsedConns    prometheus.Gauge
	PgxPoolIdleConns    prometheus.Gauge
}

// NewMetrics Создаёт экземпляры метрик из структуры
func NewMetrics() *Metrics {
	return &Metrics{
		HttpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"}),
		HttpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests",
				Buckets: []float64{0.01, 0.1, 0.5, 1, 2, 5},
			},
			[]string{"method", "path"},
		),
		PgxPoolMaxConns: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "pgxpool_max_connections",
				Help: "Maximum number of connections in the pool",
			},
		),
		PgxPoolUsedConns: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "pgxpool_used_connections",
				Help: "Currently used connections in the pool",
			},
		),
		PgxPoolIdleConns: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "pgxpool_idle_connections",
				Help: "Currently idle connections in the pool",
			},
		),
	}
}

// InitMetrics инициализируем экземпляр структуры метрик
func InitMetrics() *Metrics {
	metrics := NewMetrics()
	prometheus.MustRegister(
		metrics.HttpRequestsTotal,
		metrics.HttpRequestDuration,
		metrics.PgxPoolMaxConns,
		metrics.PgxPoolUsedConns,
		metrics.PgxPoolIdleConns,
	)
	return metrics

}
