package grpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "default-service", cfg.Service)
	assert.Equal(t, ":9090", cfg.Host)
	assert.Equal(t, 60*time.Second, cfg.MaxConnectionIdle)
	assert.Equal(t, 60*time.Second, cfg.Timeout)
	assert.Equal(t, 60*time.Second, cfg.MaxConnectionAge)
	assert.Equal(t, 60*time.Second, cfg.Time)
	assert.Equal(t, ":9091", cfg.MetricsHost)
	assert.Equal(t, 4*1024*1024, cfg.MaxRecvMsgSize)
	assert.False(t, cfg.EnableTracing)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			cfg:     DefaultConfig(),
			wantErr: false,
		},
		{
			name: "missing service",
			cfg: func() *Config {
				c := DefaultConfig()
				c.Service = ""
				return c
			}(),
			wantErr: true,
			errMsg:  "Service",
		},
		{
			name: "missing host",
			cfg: func() *Config {
				c := DefaultConfig()
				c.Host = ""
				return c
			}(),
			wantErr: true,
			errMsg:  "Host",
		},
		{
			name: "negative max recv msg size",
			cfg: func() *Config {
				c := DefaultConfig()
				c.MaxRecvMsgSize = -1
				return c
			}(),
			wantErr: true,
			errMsg:  "max_recv_msg_size must be positive",
		},
		{
			name: "negative max connection idle",
			cfg: func() *Config {
				c := DefaultConfig()
				c.MaxConnectionIdle = -1
				return c
			}(),
			wantErr: true,
			errMsg:  "max_connection_idle must be positive",
		},
		{
			name: "negative timeout",
			cfg: func() *Config {
				c := DefaultConfig()
				c.Timeout = -1
				return c
			}(),
			wantErr: true,
			errMsg:  "timeout must be positive",
		},
		{
			name: "negative max connection age",
			cfg: func() *Config {
				c := DefaultConfig()
				c.MaxConnectionAge = -1
				return c
			}(),
			wantErr: true,
			errMsg:  "max_connection_age must be positive",
		},
		{
			name: "negative time",
			cfg: func() *Config {
				c := DefaultConfig()
				c.Time = -1
				return c
			}(),
			wantErr: true,
			errMsg:  "time must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
