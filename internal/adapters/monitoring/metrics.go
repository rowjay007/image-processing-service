package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	TransformationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "image_transformations_total",
			Help: "Total number of image transformations.",
		},
		[]string{"type", "status"}, // type: sync, async; status: success, failure
	)

	QueueDepth = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "image_processing_queue_depth",
			Help: "Current depth of the image processing queue.",
		},
	)
)

func RecordRequest(method, path, status string, duration float64) {
	HttpRequestsTotal.WithLabelValues(method, path, status).Inc()
	HttpRequestDuration.WithLabelValues(method, path).Observe(duration)
}

func RecordTransformation(tType, status string) {
	TransformationsTotal.WithLabelValues(tType, status).Inc()
}

func UpdateQueueDepth(depth float64) {
	QueueDepth.Set(depth)
}
