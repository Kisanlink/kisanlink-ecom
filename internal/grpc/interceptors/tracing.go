package interceptors

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// TracingUnaryInterceptor adds OpenTelemetry distributed tracing
func TracingUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		tracer := otel.Tracer("grpc-server")
		ctx, span := tracer.Start(ctx, info.FullMethod)
		defer span.End()

		// Add attributes
		requestID := GetRequestID(ctx)
		span.SetAttributes(
			attribute.String("request_id", requestID),
			attribute.String("method", info.FullMethod),
			attribute.String("component", "grpc-server"),
		)

		// Call handler
		resp, err := handler(ctx, req)

		// Record error if present
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			span.SetAttributes(attribute.String("grpc.status_code", status.Code(err).String()))
		} else {
			span.SetStatus(codes.Ok, "success")
		}

		return resp, err
	}
}
