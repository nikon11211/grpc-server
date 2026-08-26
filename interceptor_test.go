package grpc

import (
	"context"
	"errors"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"sync"
)

type mockLogger struct {
	mu           sync.Mutex
	debugFCalled bool
	infoCalled   bool
	errorCalled  bool
}

func (m *mockLogger) DebugF(msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.debugFCalled = true
}

func (m *mockLogger) Debug(msg string) {}

func (m *mockLogger) Info(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.infoCalled = true
}

func (m *mockLogger) Warn(msg string) {}

func (m *mockLogger) Error(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorCalled = true
}

func (m *mockLogger) infoCalledFlag() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.infoCalled
}

func (m *mockLogger) errorCalledFlag() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.errorCalled
}

func TestLoggingInterceptorSuccess(t *testing.T) {
	logger := &mockLogger{}
	interceptor := loggingInterceptor(logger)

	handler := func(ctx context.Context, req any) (any, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/Service/Method"}

	resp, err := interceptor(context.Background(), "request", info, handler)
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
	assert.True(t, logger.infoCalled)
}

func TestLoggingInterceptorError(t *testing.T) {
	logger := &mockLogger{}
	interceptor := loggingInterceptor(logger)

	handler := func(ctx context.Context, req any) (any, error) {
		return nil, status.Error(codes.Internal, "test error")
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/Service/Method"}

	resp, err := interceptor(context.Background(), "request", info, handler)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, logger.infoCalled)
}

func TestMetricsInterceptorSuccess(t *testing.T) {
	interceptor := metricsInterceptor()

	handler := func(ctx context.Context, req any) (any, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/Service/Success"}

	resp, err := interceptor(context.Background(), "request", info, handler)
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
}

func TestMetricsInterceptorGRPCError(t *testing.T) {
	interceptor := metricsInterceptor()

	handler := func(ctx context.Context, req any) (any, error) {
		return nil, status.Error(codes.NotFound, "not found")
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/Service/NotFound"}

	resp, err := interceptor(context.Background(), "request", info, handler)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestMetricsInterceptorUnknownError(t *testing.T) {
	interceptor := metricsInterceptor()

	handler := func(ctx context.Context, req any) (any, error) {
		return nil, assert.AnError
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/Service/Unknown"}

	resp, err := interceptor(context.Background(), "request", info, handler)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestTracingInterceptorWithoutSpan(t *testing.T) {
	cfg := DefaultConfig()
	interceptor := tracingInterceptor(cfg)

	handler := func(ctx context.Context, req any) (any, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/Service/Traced"}

	resp, err := interceptor(context.Background(), "request", info, handler)
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
}

func TestTracingInterceptorWithRequestID(t *testing.T) {
	cfg := DefaultConfig()
	interceptor := tracingInterceptor(cfg)

	handler := func(ctx context.Context, req any) (any, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/Service/Traced"}

	resp, err := interceptor(context.Background(), "request", info, handler)
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
}

func TestExtractMethod(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/package.Service/Method", "Method"},
		{"/package.Service/GetUser", "GetUser"},
		{"MethodOnly", "MethodOnly"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractMethod(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidationInterceptorNilValidator(t *testing.T) {
	interceptor := validationInterceptor(nil)

	handler := func(ctx context.Context, req any) (any, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/Service/Method"}

	resp, err := interceptor(context.Background(), "request", info, handler)
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
}

type tracedRequest struct {
	requestID string
}

func (r *tracedRequest) GetRquid() string {
	return r.requestID
}

func TestTracingInterceptorWithSpanAndRequestID(t *testing.T) {
	cfg := DefaultConfig()
	interceptor := tracingInterceptor(cfg)

	tp := noop.NewTracerProvider()
	ctx, span := tp.Tracer("test").Start(context.Background(), "test-span")
	defer span.End()

	handler := func(ctx context.Context, req any) (any, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/Service/Traced"}

	resp, err := interceptor(ctx, &tracedRequest{requestID: "req-123"}, info, handler)
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
}

func TestTracingInterceptorWithSpanWithoutRequestID(t *testing.T) {
	cfg := DefaultConfig()
	interceptor := tracingInterceptor(cfg)

	tp := noop.NewTracerProvider()
	ctx, span := tp.Tracer("test").Start(context.Background(), "test-span")
	defer span.End()

	handler := func(ctx context.Context, req any) (any, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/Service/Traced"}

	resp, err := interceptor(ctx, "plain-request", info, handler)
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
}

type fakeValidator struct {
	err error
}

func (f *fakeValidator) Validate(msg proto.Message, _ ...protovalidate.ValidationOption) error {
	return f.err
}

func TestValidationInterceptorNonProtoMessage(t *testing.T) {
	interceptor := validationInterceptor(&fakeValidator{})

	handler := func(ctx context.Context, req any) (any, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/Service/Method"}

	resp, err := interceptor(context.Background(), "not-a-proto-message", info, handler)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestValidationInterceptorValidationError(t *testing.T) {
	interceptor := validationInterceptor(&fakeValidator{err: errors.New("violation: field is required")})

	handler := func(ctx context.Context, req any) (any, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/Service/Method"}

	resp, err := interceptor(context.Background(), &emptypb.Empty{}, info, handler)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestValidationInterceptorValid(t *testing.T) {
	interceptor := validationInterceptor(&fakeValidator{})

	handler := func(ctx context.Context, req any) (any, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/Service/Method"}

	resp, err := interceptor(context.Background(), &emptypb.Empty{}, info, handler)
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
}
