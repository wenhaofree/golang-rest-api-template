package logging

import (
	"os"
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
	return NewLoggerWithColor(levelStr, isDev, true)
}

func NewLoggerWithColor(levelStr string, isDev bool, colorized bool) (*zap.Logger, zap.AtomicLevel, error) {
	atom := zap.NewAtomicLevelAt(ParseLevel(levelStr))
	cfg := zap.Config{
		Level:       atom,
		Development: isDev,
		Encoding: func() string {
			if isDev {
				return "console"
			}
			return "json"
		}(),
		EncoderConfig: func() zapcore.EncoderConfig {
			if isDev {
				return getColoredEncoderConfig(colorized)
			}
			ec := zap.NewProductionEncoderConfig()
			ec.TimeKey = "ts"
			ec.EncodeTime = zapcore.ISO8601TimeEncoder
			return ec
		}(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		Sampling:         &zap.SamplingConfig{Initial: 100, Thereafter: 100},
	}
	logger, err := cfg.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	return logger, atom, err
}

// ANSI 颜色代码
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorGray   = "\033[37m"
	ColorCyan   = "\033[36m"
	ColorGreen  = "\033[32m"
	ColorWhite  = "\033[97m"
	ColorBold   = "\033[1m"
)

// isTerminal 检查输出是否为终端
func isTerminal(colorized bool) bool {
	if !colorized {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	// 简单检查是否为终端
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// colorizeLevel 为日志级别添加颜色
func colorizeLevel(level zapcore.Level, colorized bool) string {
	if !isTerminal(colorized) {
		return level.CapitalString()
	}

	switch level {
	case zapcore.DebugLevel:
		return ColorGray + ColorBold + level.CapitalString() + ColorReset
	case zapcore.InfoLevel:
		return ColorGreen + ColorBold + level.CapitalString() + ColorReset
	case zapcore.WarnLevel:
		return ColorYellow + ColorBold + level.CapitalString() + ColorReset
	case zapcore.ErrorLevel:
		return ColorRed + ColorBold + level.CapitalString() + ColorReset
	default:
		return ColorWhite + ColorBold + level.CapitalString() + ColorReset
	}
}

// getColoredEncoderConfig 返回支持颜色的编码器配置
func getColoredEncoderConfig(colorized bool) zapcore.EncoderConfig {
	config := zap.NewDevelopmentEncoderConfig()

	// 自定义级别编码器以支持颜色
	config.EncodeLevel = func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(colorizeLevel(level, colorized))
	}

	// 自定义时间格式
	config.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")

	// 自定义调用者信息格式
	config.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		if !isTerminal(colorized) {
			enc.AppendString(caller.TrimmedPath())
			return
		}
		enc.AppendString(ColorCyan + caller.TrimmedPath() + ColorReset)
	}

	return config
}
