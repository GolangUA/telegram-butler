package logger

import (
	"log/slog"
	"os"
)

func SetupLogger(logLevel string, logFormat string, addSource bool) *slog.Logger {
	var level slog.Level
	err := level.UnmarshalText([]byte(logLevel))
	if err != nil {
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		AddSource: addSource,
		Level:     level,
	}

	switch logFormat {
	case "json":
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	case "prettyjson":
		return slog.New(NewPrettyHandler(os.Stdout, opts))
	default:
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
}
