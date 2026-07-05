package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New builds a zap logger from the config. It defaults to JSON output with
// ISO-8601 timestamps.
func New(cfg Config) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", cfg.Level, err)
	}

	encoding, err := encoding(cfg.Format)
	if err != nil {
		return nil, err
	}

	zcfg := zap.NewProductionConfig()
	zcfg.Level = zap.NewAtomicLevelAt(level)
	zcfg.Encoding = encoding
	zcfg.EncoderConfig.TimeKey = "ts"
	zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	l, err := zcfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}

	return l, nil
}

// encoding maps the configured format onto a zap encoding name.
func encoding(format string) (string, error) {
	switch format {
	case "json":
		return "json", nil
	case "text", "console":
		return "console", nil
	default:
		return "", fmt.Errorf("invalid log format %q (want json or text)", format)
	}
}
