package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want *config
	}{
		{
			name: "default values",
			args: []string{},
			want: &config{
				port:        8443,
				certFile:    "/etc/webhook/certs/tls.crt",
				keyFile:     "/etc/webhook/certs/tls.key",
				metricsPort: 8080,
			},
		},
		{
			name: "custom values",
			args: []string{"-port=9443", "-tls-cert-file=/tmp/cert.pem", "-tls-key-file=/tmp/key.pem", "-metrics-port=9090"},
			want: &config{
				port:        9443,
				certFile:    "/tmp/cert.pem",
				keyFile:     "/tmp/key.pem",
				metricsPort: 9090,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFlagsWithArgs(tt.args)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInitLogger(t *testing.T) {
	logger, err := initLogger()
	require.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestMainInitialization(t *testing.T) {
	// This test verifies that main doesn't panic during initialization
	// In a real test, we would mock the server components
	assert.NotPanics(t, func() {
		// We can't easily test main() directly as it doesn't return
		// Instead, we test that our initialization functions work
		cfg := parseFlags()
		assert.NotNil(t, cfg)
		
		logger, err := initLogger()
		assert.NoError(t, err)
		assert.NotNil(t, logger)
	})
}