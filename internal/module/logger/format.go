package logger

import (
	"log/slog"
	"os"
)

func SetupLogger(logLevel string, logFormat string, addSource bool) *slog.Logger {
	level := StringToLevel(logLevel)

	var handler slog.Handler

	switch logFormat {
	case "text":
		// Setup for plain text formatting (assuming you have a handler for that)
		opts := &slog.HandlerOptions{
			AddSource: addSource,
			Level:     level,
		}
		handler = slog.NewTextHandler(os.Stdout, opts)
	case "json":
		// Setup for JSON formatting
		handler = slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level:     level,
				AddSource: addSource,
			},
		)
	case "prettyjson":
		opts := PrettyHandlerOptions{
			SlogOpts: slog.HandlerOptions{
				Level:     StringToLevel(logLevel),
				AddSource: addSource,
			},
		}
		handler = NewPrettyHandler(os.Stdout, opts)
	default:
		// Fallback or default logging format if none specified
		handler = slog.NewTextHandler(os.Stdout, nil)
	}

	return slog.New(handler)
}
