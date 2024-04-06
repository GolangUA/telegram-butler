//go:build local

package main

import (
	"log/slog"
	"os"

	"github.com/spf13/viper"

	"github.com/GolangUA/telegram-butler/internal/module/logger"
)

func setupLogger() *slog.Logger {
	opts := logger.PrettyHandlerOptions{
		SlogOpts: slog.HandlerOptions{
			Level: logger.StringToLevel(viper.GetString("log-level")),
		},
	}
	handler := logger.NewPrettyHandler(os.Stdout, opts)

	return slog.New(handler)
}
