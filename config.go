package grpc

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	Service           string        `mapstructure:"service" validate:"required"`
	Host              string        `mapstructure:"host" validate:"required"`
	MaxConnectionIdle time.Duration `mapstructure:"max_connection_idle" validate:"required"`
	Timeout           time.Duration `mapstructure:"timeout" validate:"required"`
	MaxConnectionAge  time.Duration `mapstructure:"max_connection_age" validate:"required"`
	Time              time.Duration `mapstructure:"time" validate:"required"`
	MetricsHost       string        `mapstructure:"metrics_host"`
	MaxRecvMsgSize    int           `mapstructure:"max_recv_msg_size" validate:"required"`
	EnableTracing     bool          `mapstructure:"enable_tracing"`
}

func DefaultConfig() *Config {
	return &Config{
		Service:           "default-service",
		Host:              ":9090",
		MaxConnectionIdle: 60 * time.Second,
		Timeout:           60 * time.Second,
		MaxConnectionAge:  60 * time.Second,
		Time:              60 * time.Second,
		MetricsHost:       ":9091",
		MaxRecvMsgSize:    4 * 1024 * 1024,
		EnableTracing:     false,
	}
}

func (c Config) Validate() error {
	v := validator.New()
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}
	if c.MaxRecvMsgSize <= 0 {
		return fmt.Errorf("max_recv_msg_size must be positive")
	}
	if c.MaxConnectionIdle <= 0 {
		return fmt.Errorf("max_connection_idle must be positive")
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	if c.MaxConnectionAge <= 0 {
		return fmt.Errorf("max_connection_age must be positive")
	}
	if c.Time <= 0 {
		return fmt.Errorf("time must be positive")
	}
	return nil
}
