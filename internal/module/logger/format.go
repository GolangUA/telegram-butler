package logger

import (
	"log/slog"
	"os"
	"strings"
)

const LevelTrace = slog.Level(-8)

func SetupLogger(logLevel string, logFormat string, addSource bool) *slog.Logger {
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
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				if a.Value.Any().(slog.Level) == LevelTrace {
					a.Value = slog.StringValue("TRACE")
				}
			}
			return a
		},
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
