package logger

import "context"

type contextKey string

const loggerContextKey contextKey = "__logger"

func ToContext(ctx context.Context, log Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, log)
}

func FromContext(ctx context.Context) Logger {
	log, ok := ctx.Value(loggerContextKey).(Logger)
	if !ok || log == nil {
		panic("no logger")
	}
	return log
}
