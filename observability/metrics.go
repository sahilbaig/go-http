package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var RequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "requests_total",
	Help: "Total HTTP Requests",
})

var RequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: prometheus.ExponentialBuckets(0.00001, 2, 5),
	},
	[]string{"path"},
)

func init() {
	prometheus.MustRegister(RequestDuration, RequestsTotal)
}
func ObserveRequest(path string, start time.Time) {
	duration := time.Since(start).Seconds()
	RequestsTotal.Inc()
	RequestDuration.WithLabelValues(path).Observe(duration)
}
