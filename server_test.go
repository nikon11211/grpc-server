package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"buf.build/go/protovalidate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

func TestNewServer(t *testing.T) {
	t.Run("with nil config", func(t *testing.T) {
		logger := &mockLogger{}
		srv, err := New(nil, logger)
		require.NoError(t, err)
		assert.NotNil(t, srv)
		srv.Stop()
	})

	t.Run("with default config", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.MetricsHost = ""
		logger := &mockLogger{}
		srv, err := New(cfg, logger)
		require.NoError(t, err)
		assert.NotNil(t, srv)
		srv.Stop()
	})

	t.Run("with custom config", func(t *testing.T) {
		cfg := &Config{
			Service:           "test-service",
			Host:              ":0",
			MaxConnectionIdle: 30 * time.Second,
			Timeout:           30 * time.Second,
			MaxConnectionAge:  30 * time.Second,
			Time:              30 * time.Second,
			MetricsHost:       "",
			MaxRecvMsgSize:    1024 * 1024,
			EnableTracing:     false,
		}
		logger := &mockLogger{}
		srv, err := New(cfg, logger)
		require.NoError(t, err)
		assert.NotNil(t, srv)
		srv.Stop()
	})

	t.Run("with options", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.MetricsHost = ""
		logger := &mockLogger{}
		srv, err := New(cfg, logger,
			WithServerOption(grpc.Creds(insecure.NewCredentials())),
			WithServerOption(grpc.MaxRecvMsgSize(2048)),
		)
		require.NoError(t, err)
		assert.NotNil(t, srv)
		srv.Stop()
	})
}

func TestServerRun(t *testing.T) {
	t.Run("valid host", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Host = ":0"
		cfg.MetricsHost = ""
		logger := &mockLogger{}
		srv, err := New(cfg, logger)
		require.NoError(t, err)

		errChan := make(chan error, 1)
		go func() {
			errChan <- srv.Run(*cfg)
		}()

		time.Sleep(100 * time.Millisecond)
		srv.Stop()

		select {
		case err := <-errChan:
			assert.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for server to stop")
		}
	})

	t.Run("invalid host", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Host = "invalid:999999"
		cfg.MetricsHost = ""
		logger := &mockLogger{}
		srv, err := New(cfg, logger)
		require.NoError(t, err)

		err = srv.Run(*cfg)
		assert.Error(t, err)
	})
}

func TestServerStop(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Host = ":0"
	cfg.MetricsHost = ""
	logger := &mockLogger{}
	srv, err := New(cfg, logger)
	require.NoError(t, err)

	go func() {
		_ = srv.Run(*cfg)
	}()

	time.Sleep(100 * time.Millisecond)
	srv.Stop()
}

func TestServerGRPCServer(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Host = ":0"
	cfg.MetricsHost = ""
	logger := &mockLogger{}
	srv, err := New(cfg, logger)
	require.NoError(t, err)

	assert.NotNil(t, srv.Server)
	srv.Stop()
}

func TestNewWithTracingEnabled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MetricsHost = ""
	cfg.EnableTracing = true
	logger := &mockLogger{}
	srv, err := New(cfg, logger, WithTracerProvider(noop.NewTracerProvider()))
	require.NoError(t, err)
	assert.NotNil(t, srv)
	assert.True(t, logger.infoCalledFlag())
	srv.Stop()
}

func TestNewWithMetricsServerError(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MetricsHost = "127.0.0.1:999999"
	logger := &mockLogger{}
	srv, err := New(cfg, logger)
	require.NoError(t, err)
	assert.NotNil(t, srv)
	require.Eventually(t, func() bool { return logger.errorCalledFlag() }, 2*time.Second, 10*time.Millisecond)
	srv.Stop()
}

func TestNewWithProtovalidateError(t *testing.T) {
	original := newValidator
	newValidator = func() (protovalidate.Validator, error) {
		return nil, errors.New("validator init failed")
	}
	t.Cleanup(func() { newValidator = original })

	cfg := DefaultConfig()
	cfg.MetricsHost = ""
	logger := &mockLogger{}
	srv, err := New(cfg, logger)
	require.NoError(t, err)
	assert.NotNil(t, srv)
	require.Eventually(t, func() bool { return logger.errorCalledFlag() }, 2*time.Second, 10*time.Millisecond)
	srv.Stop()
}

type testHealthService struct {
	grpc_health_v1.UnimplementedHealthServer
	err error
}

func (s *testHealthService) Check(ctx context.Context, _ *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
}

func newTestServer(t *testing.T, cfg *Config, logger Logger, register func(*Server), opts ...ServerOption) (*Server, string) {
	t.Helper()

	srv, err := New(cfg, logger, opts...)
	require.NoError(t, err)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := lis.Addr().String()

	register(srv)

	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Serve(lis)
	}()
	t.Cleanup(func() {
		select {
		case err := <-errChan:
			require.NoError(t, err)
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for server to stop")
		}
	})
	t.Cleanup(srv.Stop)

	return srv, addr
}
func TestServerE2E(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Host = "127.0.0.1:0"
	cfg.MetricsHost = ""
	cfg.EnableTracing = true
	logger := &mockLogger{}
	svc := &testHealthService{}
	_, addr := newTestServer(t, cfg, logger, func(s *Server) {
		grpc_health_v1.RegisterHealthServer(s.Server, svc)
	}, WithTracerProvider(noop.NewTracerProvider()))

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	client := grpc_health_v1.NewHealthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: "test"})
	require.NoError(t, err)
	assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, resp.Status)
	assert.True(t, logger.infoCalledFlag())

	svc.err = status.Error(codes.NotFound, "not found")
	_, err = client.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: "test"})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestMetricsHTTPHandler(t *testing.T) {
	metricsLis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	metricsPort := metricsLis.Addr().(*net.TCPAddr).Port
	_ = metricsLis.Close()

	cfg := DefaultConfig()
	cfg.Host = "127.0.0.1:0"
	cfg.MetricsHost = fmt.Sprintf("127.0.0.1:%d", metricsPort)
	svc := &testHealthService{}
	_, addr := newTestServer(t, cfg, &mockLogger{}, func(s *Server) {
		grpc_health_v1.RegisterHealthServer(s.Server, svc)
	})

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	client := grpc_health_v1.NewHealthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: "test"})
	require.NoError(t, err)

	svc.err = status.Error(codes.NotFound, "not found")
	_, err = client.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: "test"})
	require.Error(t, err)

	deadline := time.Now().Add(3 * time.Second)
	var body []byte
	for {
		httpResp, getErr := http.Get(fmt.Sprintf("http://127.0.0.1:%d/metrics", metricsPort))
		if getErr == nil {
			body, err = io.ReadAll(httpResp.Body)
			_ = httpResp.Body.Close()
			require.NoError(t, err)
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("metrics server did not start in time")
		}
		time.Sleep(50 * time.Millisecond)
	}

	metrics := string(body)
	assert.Contains(t, metrics, "grpc_requests_success_total")
	assert.Contains(t, metrics, "grpc_requests_error_total")
	assert.Contains(t, metrics, "grpc_request_duration_seconds")
	assert.Contains(t, metrics, "/grpc.health.v1.Health/Check")
	assert.Contains(t, metrics, `error_code="NotFound"`)
}
