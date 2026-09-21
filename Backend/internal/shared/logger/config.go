package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Build(level zapcore.Level, format string, source bool) (*zap.Logger, error) {
	logConfig := zap.NewProductionConfig()
	logConfig.Level = zap.NewAtomicLevelAt(level)
	logConfig.Encoding = format
	logConfig.DisableCaller = !source
	logConfig.OutputPaths = []string{"stdout"}
	logConfig.ErrorOutputPaths = []string{"stderr"}

	zapLog, err := logConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}
	return zapLog, nil
}
