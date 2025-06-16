package server

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewPrettyLogger creates a new zap logger with pretty (colorized) console output.
func NewPrettyLogger() *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // Colorized level
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel,
	)

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

// colorizeStatus returns a colorized string representation of the HTTP status code.
func colorizeStatus(status int) string {
	switch {
	case status >= 200 && status < 300:
		return fmt.Sprintf("\x1b[32m%d\x1b[0m", status) // Green
	case status >= 400 && status < 500:
		return fmt.Sprintf("\x1b[33m%d\x1b[0m", status) // Yellow
	case status >= 500:
		return fmt.Sprintf("\x1b[31m%d\x1b[0m", status) // Red
	default:
		return fmt.Sprintf("%d", status) // No color
	}
}
