package main

import (
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Version information - set during build
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

type config struct {
	port        int
	certFile    string
	keyFile     string
	metricsPort int
}

func main() {
	cfg := parseFlags()

	// Initialize logger
	logger, err := initLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting pod-labeler webhook",
		zap.String("version", Version),
		zap.String("commit", GitCommit),
		zap.String("built", BuildDate),
		zap.Int("port", cfg.port),
		zap.String("cert-file", cfg.certFile),
		zap.String("key-file", cfg.keyFile),
		zap.Int("metrics-port", cfg.metricsPort),
	)

	// TODO: Initialize and start webhook server
	logger.Info("Webhook server started successfully")
}

func parseFlags() *config {
	return parseFlagsWithArgs(os.Args[1:])
}

func parseFlagsWithArgs(args []string) *config {
	cfg := &config{}
	
	fs := flag.NewFlagSet("webhook", flag.ContinueOnError)
	fs.IntVar(&cfg.port, "port", 8443, "Webhook server port")
	fs.StringVar(&cfg.certFile, "tls-cert-file", "/etc/webhook/certs/tls.crt", "Path to TLS certificate file")
	fs.StringVar(&cfg.keyFile, "tls-key-file", "/etc/webhook/certs/tls.key", "Path to TLS key file")
	fs.IntVar(&cfg.metricsPort, "metrics-port", 8080, "Metrics server port")

	fs.Parse(args)

	return cfg
}

func initLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return config.Build()
}