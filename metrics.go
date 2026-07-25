package grpc

import "github.com/prometheus/client_golang/prometheus"

var histogramRequestDurationBuckets = []float64{
	0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0,
	50.0, 100.0, 1000.0, 10000.0, 100000.0,
}

func init() {
	prometheus.MustRegister(SuccessCounter, ErrorCounter, RequestDuration)
}

var (
	SuccessCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_success_total",
			Help: "Total number of successful gRPC requests",
		},
		[]string{"method"},
	)

	ErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_error_total",
			Help: "Total number of failed gRPC requests",
		},
		[]string{"method", "error_code"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Histogram of latencies for gRPC requests",
			Buckets: histogramRequestDurationBuckets,
		},
		[]string{"method"},
	)
)
