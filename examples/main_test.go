package main

import (
	"testing"
	"time"

	grpcserver "github.com/nikon11211/grpc-server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestExampleServerCreation(t *testing.T) {
	cfg := grpcserver.DefaultConfig()
	cfg.Host = ":0"
	cfg.MetricsHost = ""

	logger := &mockLogger{}
	srv, err := grpcserver.New(cfg, logger,
		grpcserver.WithServerOption(grpc.MaxConcurrentStreams(100)),
	)
	require.NoError(t, err)
	assert.NotNil(t, srv)
	srv.Stop()
}

func TestExampleServerRun(t *testing.T) {
	cfg := grpcserver.DefaultConfig()
	cfg.Host = ":0"
	cfg.MetricsHost = ""

	logger := &mockLogger{}
	srv, err := grpcserver.New(cfg, logger)
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
		t.Fatal("timeout")
	}
}

type mockLogger struct{}

func (m mockLogger) DebugF(msg string, args ...any) {}
func (m mockLogger) Debug(msg string)               {}
func (m mockLogger) Info(msg string)                {}
func (m mockLogger) Warn(msg string)                {}
func (m mockLogger) Error(msg string)               {}
