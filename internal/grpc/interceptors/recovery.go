// Package interceptors provides gRPC interceptors for the server
package interceptors

import (
	"context"
	"runtime/debug"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryUnaryInterceptor recovers from panics and returns internal error to client
func RecoveryUnaryInterceptor(logger *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				logger.WithFields(logrus.Fields{
					"panic":  r,
					"method": info.FullMethod,
					"stack":  stack,
				}).Error("Panic recovered in gRPC handler")

				// Return internal error to client (don't leak panic details)
				err = status.Errorf(codes.Internal, "Internal server error")
			}
		}()

		resp, err = handler(ctx, req)
		return resp, err
	}
}
