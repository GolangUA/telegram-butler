package logger

import (
	"log/slog"
	"os"
)

func SetupLogger(logLevel string, logFormat string, addSource bool) *slog.Logger {
	level := StringToLevel(logLevel)
	var (
		handler  slog.Handler
		replAttr func(groups []string, a slog.Attr) slog.Attr
	)

	// TODO hide token values
	// replAttr = func(_ []string, a slog.Attr) slog.Attr {
	// 	if a.Value
	// 	return a
	// }

	switch logFormat {
	case "text":
		// Setup for plain text formatting (assuming you have a handler for that)
		opts := &slog.HandlerOptions{
			AddSource:   addSource,
			Level:       level,
			ReplaceAttr: replAttr,
		}
		handler = slog.NewTextHandler(os.Stdout, opts)
	case "json":
		// Setup for JSON formatting
		handler = slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level:       level,
				AddSource:   addSource,
				ReplaceAttr: replAttr,
			},
		)
	case "prettyjson":
		opts := PrettyHandlerOptions{
			SlogOpts: slog.HandlerOptions{
				Level:       StringToLevel(logLevel),
				AddSource:   addSource,
				ReplaceAttr: replAttr,
			},
		}
		handler = NewPrettyHandler(os.Stdout, opts)
	default:
		// Fallback or default logging format if none specified
		handler = slog.NewTextHandler(os.Stdout, nil)
	}

	return slog.New(handler)
}
