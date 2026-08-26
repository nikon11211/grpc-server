package grpc

import (
	"context"
	"testing"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	grsBenchAny any
	grsBenchErr error
	grsBenchStr string
)

type benchNoopLogger struct{}

func (benchNoopLogger) DebugF(msg string, args ...any) {}
func (benchNoopLogger) Debug(msg string)               {}
func (benchNoopLogger) Info(msg string)                {}
func (benchNoopLogger) Warn(msg string)                {}
func (benchNoopLogger) Error(msg string)               {}

type benchValidator struct{}

func (benchValidator) Validate(_ proto.Message, _ ...protovalidate.ValidationOption) error {
	return nil
}

func BenchmarkExtractMethod(b *testing.B) {
	for i := 0; i < b.N; i++ {
		grsBenchStr = extractMethod("/app.v1.OrderService/GetByID")
	}
}

func BenchmarkConfigValidate(b *testing.B) {
	cfg := DefaultConfig()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		grsBenchErr = cfg.Validate()
	}
}

func BenchmarkServerNew(b *testing.B) {
	cfg := DefaultConfig()
	cfg.MetricsHost = ""
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		grsBenchAny, grsBenchErr = New(cfg, benchNoopLogger{})
	}
}

func BenchmarkServerNewWithOptions(b *testing.B) {
	cfg := DefaultConfig()
	cfg.MetricsHost = ""
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		grsBenchAny, grsBenchErr = New(cfg, benchNoopLogger{},
			WithServerOption(grpc.MaxConcurrentStreams(100)),
		)
	}
}

func BenchmarkLoggingInterceptor(b *testing.B) {
	interceptor := loggingInterceptor(benchNoopLogger{})
	info := &grpc.UnaryServerInfo{FullMethod: "/app.v1.OrderService/GetByID"}
	handler := func(ctx context.Context, req any) (any, error) { return req, nil }
	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		grsBenchAny, grsBenchErr = interceptor(ctx, "request", info, handler)
	}
}

func BenchmarkMetricsInterceptor(b *testing.B) {
	interceptor := metricsInterceptor()
	info := &grpc.UnaryServerInfo{FullMethod: "/app.v1.OrderService/Create"}
	handler := func(ctx context.Context, req any) (any, error) { return req, nil }
	ctx := context.Background()
	req := &benchRequest{name: "bench", rquid: "request-id"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		grsBenchAny, grsBenchErr = interceptor(ctx, req, info, handler)
	}
}

func BenchmarkValidationInterceptor(b *testing.B) {
	var validator protovalidate.Validator = benchValidator{}
	interceptor := validationInterceptor(validator)
	info := &grpc.UnaryServerInfo{FullMethod: "/app.v1.OrderService/Create"}
	handler := func(ctx context.Context, req any) (any, error) { return req, nil }
	ctx := context.Background()
	req := &emptypb.Empty{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		grsBenchAny, grsBenchErr = interceptor(ctx, req, info, handler)
	}
}

func BenchmarkTracingInterceptor(b *testing.B) {
	cfg := DefaultConfig()
	interceptor := tracingInterceptor(cfg)
	info := &grpc.UnaryServerInfo{FullMethod: "/app.v1.OrderService/GetByID"}
	handler := func(ctx context.Context, req any) (any, error) { return req, nil }
	ctx := context.Background()
	req := &benchRequest{name: "bench", rquid: "request-id"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		grsBenchAny, grsBenchErr = interceptor(ctx, req, info, handler)
	}
}

type benchRequest struct {
	name  string
	rquid string
}

func (r *benchRequest) GetName() string { return r.name }

func (r *benchRequest) GetRquid() string { return r.rquid }
