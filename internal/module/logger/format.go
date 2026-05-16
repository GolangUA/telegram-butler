package logger

import (
	"log/slog"
	"os"
	"strings"
)

const LevelTrace = slog.Level(-8)

func Setup(logLevel string, logFormat string, addSource bool) *slog.Logger {
	var level slog.Level
	if strings.EqualFold(logLevel, "trace") {
		level = LevelTrace
	} else {
		err := level.UnmarshalText([]byte(logLevel))
		if err != nil {
			level = slog.LevelInfo
		}
	}

	opts := &slog.HandlerOptions{
		AddSource: addSource,
		Level:     level,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				level, ok := a.Value.Any().(slog.Level)
				if ok && level == LevelTrace {
					a.Value = slog.StringValue("TRACE")
				}
			}

			return a
		},
	}

	var logger *slog.Logger

	switch logFormat {
	case "json":
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	case "prettyjson":
		logger = slog.New(NewPrettyHandler(os.Stdout, opts))
	default:
		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	}

	// Route the std `log` package through slog so library writes
	// (ngrok agent proxy errors, etc.) use the same handler/format.
	slog.SetDefault(logger)

	return logger
}
