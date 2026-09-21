package logger

import (
	"context"

	"go.uber.org/zap"
)

type zapLogger struct {
	log *zap.SugaredLogger
}

func NewZap(log *zap.Logger) Logger {
	if log == nil {
		return nil
	}
	return &zapLogger{log: log.Sugar()}
}

func (l *zapLogger) Debug(_ context.Context, msg string, args ...any) {
	l.log.Debugw(msg, args...)
}

func (l *zapLogger) Info(_ context.Context, msg string, args ...any) {
	l.log.Infow(msg, args...)
}

func (l *zapLogger) Warn(_ context.Context, msg string, args ...any) {
	l.log.Warnw(msg, args...)
}

func (l *zapLogger) Error(_ context.Context, msg string, args ...any) {
	l.log.Errorw(msg, args...)
}

func (l *zapLogger) With(args ...any) Logger {
	return &zapLogger{log: l.log.With(args...)}
}
