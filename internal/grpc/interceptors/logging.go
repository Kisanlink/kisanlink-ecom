package interceptors

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingUnaryInterceptor logs all gRPC requests and responses
func LoggingUnaryInterceptor(logger *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		requestID := GetRequestID(ctx)

		logger.WithFields(logrus.Fields{
			"method":     info.FullMethod,
			"request_id": requestID,
		}).Info("gRPC request started")

		// Call handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start)

		// Determine status
		code := status.Code(err)
		logEntry := logger.WithFields(logrus.Fields{
			"method":     info.FullMethod,
			"request_id": requestID,
			"duration":   duration,
			"status":     code.String(),
		})

		if err != nil {
			logEntry.WithError(err).Error("gRPC request failed")
		} else {
			logEntry.Info("gRPC request completed")
		}

		return resp, err
	}
}
