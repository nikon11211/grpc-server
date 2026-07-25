package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockLogger struct {
	debugFCalled bool
	infoCalled   bool
	errorCalled  bool
}

func (m *mockLogger) DebugF(msg string, args ...any) {
	m.debugFCalled = true
}

func (m *mockLogger) Debug(msg string) {}

func (m *mockLogger) Info(msg string) {
	m.infoCalled = true
}

func (m *mockLogger) Warn(msg string) {}

func (m *mockLogger) Error(msg string) {
	m.errorCalled = true
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
