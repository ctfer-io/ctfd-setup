package ctfdsetup

import (
	"sync"

	"go.uber.org/zap"
)

var (
	logSync sync.Once
	logger  *zap.Logger
)

func Log() *zap.Logger {
	logSync.Do(func() {
		logger, _ = zap.NewProduction()
	})
	return logger
}
