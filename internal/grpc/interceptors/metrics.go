package interceptors

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	grpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method"},
	)

	grpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Duration of gRPC requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	grpcRequestsByStatus = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_by_status_total",
			Help: "Total number of gRPC requests by status code",
		},
		[]string{"method", "status"},
	)
)

// MetricsUnaryInterceptor collects Prometheus metrics for gRPC requests
func MetricsUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Increment request counter
		grpcRequestsTotal.WithLabelValues(info.FullMethod).Inc()

		// Call handler
		resp, err := handler(ctx, req)

		// Record duration
		duration := time.Since(start).Seconds()
		grpcRequestDuration.WithLabelValues(info.FullMethod).Observe(duration)

		// Record status
		code := status.Code(err)
		grpcRequestsByStatus.WithLabelValues(info.FullMethod, code.String()).Inc()

		return resp, err
	}
}
