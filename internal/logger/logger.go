package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log   *zap.Logger        = zap.NewNop()
	Sugar *zap.SugaredLogger = zap.NewNop().Sugar()
)

func NewZapLogger(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	cfg.DisableStacktrace = true
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	Sugar = zl.Sugar()
	return nil
}
