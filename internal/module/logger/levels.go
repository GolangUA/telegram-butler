package logger

import (
	"log/slog"
	"strings"
)

/*
StringToLevel converts a string representation of a log level into its corresponding slog.Level.

LevelDebug Level = -4
LevelInfo  Level = 0
LevelWarn  Level = 4
LevelError Level = 8
*/
func StringToLevel(s string) slog.Level {
	s = strings.ToLower(s)
	switch s {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
