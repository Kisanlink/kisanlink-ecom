package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Validator interface for proto message validation
type Validator interface {
	Validate() error
}

// ValidationUnaryInterceptor validates proto messages if they implement Validate() method
func ValidationUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Check if request implements Validate() method
		if validator, ok := req.(Validator); ok {
			if err := validator.Validate(); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "Validation failed: %v", err)
			}
		}

		return handler(ctx, req)
	}
}
