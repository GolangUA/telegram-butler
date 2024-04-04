package logger

import (
	"context"
	"log/slog"
)

type contextKey string

const loggerContextKey contextKey = "__logger"

func ToContext(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, log)
}

func FromContext(ctx context.Context) *slog.Logger {
	log, ok := ctx.Value(loggerContextKey).(*slog.Logger)
	if !ok || log == nil {
		panic("no logger")
	}
	return log
}
