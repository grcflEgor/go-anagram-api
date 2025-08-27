package logger

import (
	"context"

	"go.uber.org/zap"
)

type contextKey string

const loggerKey = contextKey("logger")

func WithLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

func FromContext(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return l
	}
	return AppLogger
}
