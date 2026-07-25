package grpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"buf.build/go/protovalidate"
	"go.opentelemetry.io/otel/attribute"
	ottrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func loggingInterceptor(logger Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		startTime := time.Now()

		resp, err := handler(ctx, req)

		st := "OK"
		if err != nil {
			st = "ERROR"
		}

		logger.Info(fmt.Sprintf("%s %s %v", info.FullMethod, st, time.Since(startTime)))
		return resp, err
	}
}

func metricsInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		startTime := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(startTime).Seconds()
		method := info.FullMethod

		RequestDuration.WithLabelValues(method).Observe(duration)

		if err != nil {
			errorCode := "unknown"
			if grpcErr, ok := status.FromError(err); ok {
				errorCode = grpcErr.Code().String()
			}
			ErrorCounter.WithLabelValues(method, errorCode).Inc()
		} else {
			SuccessCounter.WithLabelValues(method).Inc()
		}

		return resp, err
	}
}

func tracingInterceptor(cfg *Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var requestID string

		if r, ok := req.(interface{ GetRquid() string }); ok {
			requestID = r.GetRquid()
		}

		resp, err := handler(ctx, req)

		span := ottrace.SpanFromContext(ctx)
		if span != nil {
			attrs := []attribute.KeyValue{
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.service", info.FullMethod),
				attribute.String("rpc.method", extractMethod(info.FullMethod)),
				attribute.String("net.protocol", "grpc"),
				attribute.String("grpc.server", cfg.Service),
				attribute.String("grpc.host", cfg.Host),
			}

			if requestID != "" {
				attrs = append(attrs, attribute.String("x-request-id", requestID))
			}

			span.SetAttributes(attrs...)
		}

		return resp, err
	}
}

func extractMethod(fullMethod string) string {
	parts := strings.Split(fullMethod, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullMethod
}

func validationInterceptor(validator protovalidate.Validator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if validator == nil {
			return handler(ctx, req)
		}

		msg, ok := req.(proto.Message)
		if !ok {
			return nil, status.Error(codes.Internal, "request does not implement proto.Message")
		}

		if err := validator.Validate(msg); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return handler(ctx, req)
	}
}
