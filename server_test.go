package grpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
