package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bxf1/ERP/backend/config"
)

var L *zap.Logger

func Init(cfg config.LogConfig) error {
	var zcfg zap.Config
	if cfg.Format == "json" {
		zcfg = zap.NewProductionConfig()
	} else {
		zcfg = zap.NewDevelopmentConfig()
	}

	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	zcfg.Level = zap.NewAtomicLevelAt(level)
	zcfg.EncoderConfig.TimeKey = "timestamp"
	zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	L, err = zcfg.Build()
	return err
}

func Sync() {
	if L != nil {
		_ = L.Sync()
	}
}

// WithTenant returns a logger with tenant_id field.
func WithTenant(tenantID string) *zap.Logger {
	return L.With(zap.String("tenant_id", tenantID))
}

// WithRequest returns a logger with request-scoped fields.
func WithRequest(tenantID, traceID string) *zap.Logger {
	return L.With(
		zap.String("tenant_id", tenantID),
		zap.String("trace_id", traceID),
	)
}
