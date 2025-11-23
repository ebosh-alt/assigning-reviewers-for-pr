// Package logger provides zap logger initialization.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New constructs a sugared zap logger with given level.
func New(level string) (*zap.SugaredLogger, error) {
	var zapLevel zapcore.Level
	err := zapLevel.UnmarshalText([]byte(level))
	if err != nil {
		return nil, err
	}

	config := zap.Config{
		Encoding:         "console", // "console" — для красивого CLI-лога
		Level:            zap.NewAtomicLevelAt(zapLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		DisableStacktrace: true,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}

	log, err := config.Build()
	if err != nil {
		return nil, err
	}
	return log.Sugar(), nil
}
