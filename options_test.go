package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestWithTracerProvider(t *testing.T) {
	opts := &serverOptions{}
	tp := noop.NewTracerProvider()

	opt := WithTracerProvider(tp)
	opt(opts)

	assert.NotNil(t, opts.tracerProvider)
	assert.Equal(t, tp, opts.tracerProvider)
}

func TestWithServerOption(t *testing.T) {
	opts := &serverOptions{}

	grpcOpt := grpc.Creds(insecure.NewCredentials())
	opt := WithServerOption(grpcOpt)
	opt(opts)

	assert.Len(t, opts.serverOptions, 1)
}

func TestMultipleOptions(t *testing.T) {
	opts := &serverOptions{}
	tp := noop.NewTracerProvider()

	WithTracerProvider(tp)(opts)
	WithServerOption(grpc.MaxRecvMsgSize(1024))(opts)
	WithServerOption(grpc.MaxConcurrentStreams(100))(opts)

	assert.NotNil(t, opts.tracerProvider)
	assert.Len(t, opts.serverOptions, 2)
}
