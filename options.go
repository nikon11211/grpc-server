package grpc

import (
	ottrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type ServerOption func(*serverOptions)

type serverOptions struct {
	tracerProvider ottrace.TracerProvider
	serverOptions  []grpc.ServerOption
}

func WithTracerProvider(tp ottrace.TracerProvider) ServerOption {
	return func(o *serverOptions) {
		o.tracerProvider = tp
	}
}

func WithServerOption(opt grpc.ServerOption) ServerOption {
	return func(o *serverOptions) {
		o.serverOptions = append(o.serverOptions, opt)
	}
}
