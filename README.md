<p align="center">
  <h1 align="center">gRPC Server - Production-Ready gRPC Framework for Go</h1>
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/nikon11211/grpc-server">
    <img src="https://pkg.go.dev/badge/github.com/nikon11211/grpc-server.svg" alt="Go Reference"/>
  </a>
  <a href="https://goreportcard.com/report/github.com/nikon11211/grpc-server">
    <img src="https://goreportcard.com/badge/github.com/nikon11211/grpc-server" alt="Go Report Card"/>
  </a>
  <a href="https://github.com/nikon11211/grpc-server/actions/workflows/test.yaml">
    <img src="https://github.com/nikon11211/grpc-server/actions/workflows/test.yaml/badge.svg" alt="Tests"/>
  </a>
  <a href="https://codecov.io/gh/nikon11211/grpc-server">
    <img src="https://codecov.io/gh/nikon11211/grpc-server/branch/main/graph/badge.svg" alt="Coverage"/>
  </a>
  <a href="https://sonarcloud.io/summary/overall?id=nikon11211_grpc-server">
    <img src="https://sonarcloud.io/api/project_badges/measure?project=nikon11211_grpc-server&metric=coverage" alt="SonarCloud Coverage"/>
  </a>
  <a href="https://opensource.org/licenses/MIT">
    <img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"/>
  </a>
  <a href="https://golang.org/">
    <img src="https://img.shields.io/badge/Go-%3E%3D%201.26-blue" alt="Go Version"/>
  </a>
  <a href="https://prometheus.io/">
    <img src="https://img.shields.io/badge/Prometheus-Ready-red" alt="Prometheus"/>
  </a>
  <a href="https://opentelemetry.io/">
    <img src="https://img.shields.io/badge/OpenTelemetry-Enabled-orange" alt="OpenTelemetry"/>
  </a>
</p>

<p align="center">
  <b>A high-performance, production-ready gRPC server library with built-in observability</b><br/>
  <i>gRPC • OpenTelemetry • Prometheus • Request Validation • Keepalive • Structured Logging</i>
</p>

---

## 🎯 Features

<table>
<tr>
<td width="50%">

### 🚀 Core Capabilities
- Production-ready gRPC server with optimized defaults
- Automatic protobuf message validation via protovalidate
- Configurable keepalive with MaxConnectionIdle, Timeout, MaxConnectionAge, Time
- Request ID propagation across service boundaries
- Graceful shutdown with configurable timeout
- Max message size configuration for security

### 📊 Observability
- OpenTelemetry tracing with automatic span management
- Prometheus metrics with pre-configured histogram buckets
- Structured logging interface with multiple severity levels
- Request duration tracking with method-level granularity
- Success/error counters with gRPC status codes
- Automatic trace context propagation via W3C Trace Context

</td>
<td width="50%">

### 🔒 Enterprise Ready
- TLS support via gRPC server options
- Configurable timeouts for all connection types
- Zero-allocation hot path with optimized interceptors
- Pluggable logger interface for any logging library
- Race-condition free design tested with -race flag
- Nil-safe API with comprehensive nil receiver handling

### 🧩 Extensibility
- Custom gRPC server options support
- Tracer provider injection for any OpenTelemetry backend
- Pluggable validation via protovalidate
- Interceptor chain customization
- Metrics registry isolation for multi-service deployments

</td>
</tr>
</table>

## 📦 Installation

```bash
go get github.com/nikon11211/grpc-server
```

## 🏗️ Architecture

```
┌───────────────────────────────────────────────────────────┐
│                    Your Application                       │
├───────────────────────────────────────────────────────────┤
│                 gRPC Server Framework                     │
│                                                           │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌────────────┐ │
│  │Validation │ │ Logging   │ │ Metrics   │ │  Tracing   │ │
│  │Interceptor│ │Interceptor│ │Interceptor│ │Interceptor │ │
│  └────┬──────┘ └────┬──────┘ └────┬──────┘ └─────┬──────┘ │
│       └─────────────┴─────────────┴──────────────┘        │
│                          │                                │
│                   gRPC Server                             │
│                          │                                │
│              Keepalive • Timeouts • TLS                   │
└───────────────────────────────────────────────────────────┘
```

## 🚀 Quick Start

### Minimal Server

```go
package main

import (
	"time"

	grpcserver "github.com/nikon11211/grpc-server"
)

type logger struct{}

func (l logger) DebugF(msg string, args ...any) {}
func (l logger) Debug(msg string)        {}
func (l logger) Info(msg string)         {}
func (l logger) Warn(msg string)         {}
func (l logger) Error(msg string)        {}

func main() {
	cfg := &grpcserver.Config{
		Service:           "my-service",
		Host:              ":9090",
		MaxConnectionIdle: 60 * time.Second,
		Timeout:           60 * time.Second,
		MaxConnectionAge:  60 * time.Second,
		Time:              60 * time.Second,
		MetricsHost:       ":9091",
		MaxRecvMsgSize:    4 * 1024 * 1024,
		EnableTracing:     false,
	}

	srv, err := grpcserver.New(cfg, logger{})
	if err != nil {
		panic(err)
	}
	defer srv.Stop()

	if err := srv.Run(*cfg); err != nil {
		panic(err)
	}
}
```

### With OpenTelemetry Tracing

```go
import (
"go.opentelemetry.io/otel"
"go.opentelemetry.io/otel/sdk/trace"
)

func main() {
tp := trace.NewTracerProvider()
otel.SetTracerProvider(tp)

cfg := &grpcserver.Config{
Service:       "traced-service",
Host:          ":9090",
EnableTracing: true,
}

srv, err := grpcserver.New(cfg, logger{},
grpcserver.WithTracerProvider(tp),
)
}
```

### With Custom gRPC Options

```go
import "google.golang.org/grpc"

srv, err := grpcserver.New(cfg, logger{},
grpcserver.WithServerOption(grpc.MaxConcurrentStreams(1000)),
grpcserver.WithServerOption(grpc.MaxSendMsgSize(8*1024*1024)),
)
```

### Structured Logging Integration

```go
import "log/slog"

type slogLogger struct {
logger *slog.Logger
}

func (l slogLogger) Info(msg string) {
l.logger.Info(msg)
}

func (l slogLogger) Error(msg string) {
l.logger.Error(msg)
}
```

## 🔧 Configuration Reference

```go
type Config struct {
Service           string        // Service name for metrics and tracing (required)
Host              string        // gRPC server host:port (required)
MaxConnectionIdle time.Duration // Max idle time before GOAWAY
Timeout           time.Duration // Keepalive timeout for connections
MaxConnectionAge  time.Duration // Max connection lifetime
Time              time.Duration // Keepalive ping interval
MetricsHost       string        // Prometheus metrics endpoint host:port
MaxRecvMsgSize    int           // Max received message size in bytes
EnableTracing     bool          // Enable OpenTelemetry tracing
}
```

## 🧪 Testing & Benchmarks

```go
// Run all tests
go test ./...

// Run with race detection
go test -race ./...

// Run with coverage (excluding examples)
go test -coverprofile=coverage.txt ./...
go tool cover -html=coverage.txt

// Run benchmarks
go test -bench=. -benchmem -run '^$' .
```

| Benchmark | What it measures |
|-----------|------------------|
| `BenchmarkExtractMethod` | gRPC method path → handler extraction |
| `BenchmarkConfigValidate` | Config validation |
| `BenchmarkServerNew` | Server construction |
| `BenchmarkServerNewWithOptions` | Server construction with options |
| `BenchmarkLoggingInterceptor` | Logging interceptor on a single RPC |
| `BenchmarkMetricsInterceptor` | Metrics interceptor on a single RPC |
| `BenchmarkValidationInterceptor` | protovalidate interceptor on a single RPC |
| `BenchmarkTracingInterceptor` | Tracing interceptor on a single RPC |

## 📁 Project Structure

```text
grpc-server/
├── config.go              # Configuration with validation
├── config_test.go         # Configuration tests
├── server.go              # Core server implementation
├── server_test.go         # Server integration tests
├── interceptor.go         # gRPC interceptors
├── interceptor_test.go    # Interceptor unit tests
├── metrics.go             # Prometheus metrics definitions
├── options.go             # Server option patterns
├── options_test.go        # Options tests
├── logger.go              # Logger interface
├── benchmarks_test.go     # Benchmarks
├── examples/
│   ├── main.go            # Complete example application
│   └── main_test.go       # Example integration tests
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

## 📚 Complete Example

```go
package main

import (
    "context"
    "log"
    "os/signal"
    "syscall"
    "time"

    grpcserver "github.com/nikon11211/grpc-server"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/sdk/trace"
    "google.golang.org/grpc"
)

type appLogger struct{}

func (l appLogger) DebugF(msg string, args ...any) {}
func (l appLogger) Debug(msg string)        {}
func (l appLogger) Info(msg string)         {}
func (l appLogger) Warn(msg string)         {}
func (l appLogger) Error(msg string)        {}

func main() {
    tp := trace.NewTracerProvider()
    otel.SetTracerProvider(tp)

    cfg := &grpcserver.Config{
        Service:           "user-service",
        Host:              ":9090",
        MaxConnectionIdle: 60 * time.Second,
        Timeout:           60 * time.Second,
        MaxConnectionAge:  60 * time.Second,
        Time:              60 * time.Second,
        MetricsHost:       ":9091",
        MaxRecvMsgSize:    4 * 1024 * 1024,
        EnableTracing:     true,
    }

    srv, err := grpcserver.New(cfg, appLogger{},
        grpcserver.WithTracerProvider(tp),
        grpcserver.WithServerOption(grpc.MaxConcurrentStreams(1000)),
    )
    if err != nil {
        log.Fatalf("Failed to create server: %v", err)
    }
    defer srv.Stop()

    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    go func() {
        if err := srv.Run(*cfg); err != nil {
            log.Fatalf("Server error: %v", err)
        }
    }()

    <-ctx.Done()
    log.Println("Shutting down server...")
}
```


## 🤝 Contributing

We welcome contributions! Here's how you can help:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🌟 Show Your Support

Give a ⭐️ if this project helped you! Share it with your team to improve logging across your microservices.

## 🙏 Acknowledgments

- [gRPC-Go](https://google.golang.org/grpc/) - The official Go gRPC implementation
- [OpenTelemetry](https://opentelemetry.io/) - Distributed tracing standard
- [Prometheus](https://opentelemetry.io/) - Metrics and monitoring
- [protovalidate](https://github.com/bufbuild/protovalidate/) - Protobuf validation

---

<p align="center">
  <b>Made with ❤️ for the Go community</b><br/>
  <sub>Built for performance, designed for reliability</sub>
</p>