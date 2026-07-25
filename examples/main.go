package main

import (
	grpcserver "github.com/nikon11211/grpc-server"
	"google.golang.org/grpc"
)

type appLogger struct{}

func (l appLogger) DebugF(msg string, args ...any) {}
func (l appLogger) Debug(msg string)               {}
func (l appLogger) Info(msg string)                {}
func (l appLogger) Warn(msg string)                {}
func (l appLogger) Error(msg string)               {}

func main() {
	cfg := grpcserver.DefaultConfig()
	cfg.Service = "example-service"
	cfg.MetricsHost = ""

	srv, err := grpcserver.New(cfg, appLogger{},
		grpcserver.WithServerOption(grpc.MaxConcurrentStreams(1000)),
	)
	if err != nil {
		panic(err)
	}
	defer srv.Stop()

	if err := srv.Run(*cfg); err != nil {
		panic(err)
	}
}
