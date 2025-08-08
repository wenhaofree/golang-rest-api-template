package logging

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ParseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func NewLogger(levelStr string, isDev bool) (*zap.Logger, zap.AtomicLevel, error) {
	atom := zap.NewAtomicLevelAt(ParseLevel(levelStr))
	cfg := zap.Config{
		Level:       atom,
		Development: isDev,
		Encoding:    func() string { if isDev { return "console" }; return "json" }(),
		EncoderConfig: func() zapcore.EncoderConfig {
			if isDev { return zap.NewDevelopmentEncoderConfig() }
			ec := zap.NewProductionEncoderConfig()
			ec.TimeKey = "ts"
			ec.EncodeTime = zapcore.ISO8601TimeEncoder
			return ec
		}(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		Sampling: &zap.SamplingConfig{Initial: 100, Thereafter: 100},
	}
	logger, err := cfg.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	return logger, atom, err
}

