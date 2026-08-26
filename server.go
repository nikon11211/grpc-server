package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"buf.build/go/protovalidate"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type Server struct {
	Logger
	*grpc.Server
	metricsServer *http.Server
}

var newValidator = func() (protovalidate.Validator, error) {
	return protovalidate.New()
}

func New(cfg *Config, logger Logger, opts ...ServerOption) (*Server, error) {
	const trace = "grpc.New"

	if cfg == nil {
		cfg = DefaultConfig()
	}

	options := &serverOptions{}
	for _, opt := range opts {
		opt(options)
	}

	var metricsServer *http.Server
	if cfg.MetricsHost != "" {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		metricsServer = &http.Server{
			Addr:              cfg.MetricsHost,
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
		}
		go func() {
			if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error(fmt.Sprintf("(%s) metrics server error: %v", trace, err))
			}
		}()
	}

	validator, err := newValidator()
	if err != nil {
		logger.Error(fmt.Sprintf("(%s) protovalidate initialization error: %v", trace, err))
	}

	serverOpts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: cfg.MaxConnectionIdle,
			Timeout:           cfg.Timeout,
			MaxConnectionAge:  cfg.MaxConnectionAge,
			Time:              cfg.Time,
		}),
		grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSize),
	}

	serverOpts = append(serverOpts, options.serverOptions...)

	interceptors := []grpc.UnaryServerInterceptor{
		loggingInterceptor(logger),
		metricsInterceptor(),
	}

	if validator != nil {
		interceptors = append([]grpc.UnaryServerInterceptor{validationInterceptor(validator)}, interceptors...)
	}

	if cfg.EnableTracing && options.tracerProvider != nil {
		interceptors = append(interceptors, tracingInterceptor(cfg))
		serverOpts = append(serverOpts,
			grpc.StatsHandler(otelgrpc.NewServerHandler(
				otelgrpc.WithTracerProvider(options.tracerProvider),
				otelgrpc.WithPropagators(propagation.TraceContext{}),
			)),
		)
		logger.Info(fmt.Sprintf("(%s) tracing enabled", trace))
	}

	serverOpts = append(serverOpts, grpc.ChainUnaryInterceptor(interceptors...))

	srv := grpc.NewServer(serverOpts...)

	return &Server{
		Logger:        logger,
		Server:        srv,
		metricsServer: metricsServer,
	}, nil
}

func (s *Server) Run(cfg Config) error {
	const trace = "grpc.Server.Run"

	lis, err := net.Listen("tcp", cfg.Host)
	if err != nil {
		return fmt.Errorf("(%s) listener creation error: %w", trace, err)
	}

	s.Info(fmt.Sprintf("(%s) server started on %s", trace, cfg.Host))
	return s.Serve(lis)
}

func (s *Server) Stop() {
	s.GracefulStop()
	if s.metricsServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.metricsServer.Shutdown(ctx)
	}
}
