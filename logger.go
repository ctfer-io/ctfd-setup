package ctfdsetup

import (
	"context"
	"os"
	"sync"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	sub *zap.Logger
}

func (log *Logger) Info(_ context.Context, msg string, fields ...zap.Field) {
	log.sub.Info(msg, fields...)
}

func (log *Logger) Debug(_ context.Context, msg string, fields ...zap.Field) {
	log.sub.Debug(msg, fields...)
}

func (log *Logger) Error(_ context.Context, msg string, fields ...zap.Field) {
	log.sub.Error(msg, fields...)
}

var (
	loggerMx sync.Mutex
	logger   *Logger
)

// Log returns the zap logger, ready to use.
// It exports no log, and is configured at level "info".
func Log() *Logger {
	loggerMx.Lock()
	defer loggerMx.Unlock()

	// If used as a library, defaults to the most basic logger we can get
	if logger != nil {
		return logger
	}
	sub, _ := zap.NewProduction()
	logger = &Logger{sub: sub}
	return logger
}

// UpsertLogger overrides the global logger used by the tool for future operations.
//
// Ideally, it should be called once at the beginning our your tool integration, but
// you can adapt to match your needs (e.g., hot level-reconfiguration).
func UpsertLogger(prov log.LoggerProvider, level string) *Logger {
	loggerMx.Lock()
	defer loggerMx.Unlock()

	lvl, _ := zapcore.ParseLevel(level)
	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.AddSync(os.Stdout),
			lvl,
		),
		otelzap.NewCore(
			serviceName,
			otelzap.WithLoggerProvider(prov),
		),
	)
	logger = &Logger{sub: zap.New(core)}
	return logger
}
