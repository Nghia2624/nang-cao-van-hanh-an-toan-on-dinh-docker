package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.SugaredLogger
}

func New(level string) *Logger {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "json"
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.LevelKey = "level"
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.MessageKey = "msg"

	if lv, err := zap.ParseAtomicLevel(level); err == nil {
		cfg.Level = lv
	} else {
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	z, _ := cfg.Build()
	return &Logger{z.Sugar()}
}

// Field is a helper to keep imports small at call sites.
func Field(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}
