package main

import (
	"log/slog"
	"os"

	"github.com/spf13/viper"

	"github.com/GolangUA/telegram-butler/internal/module/logger"
)

func setupLogger() *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level:     logger.StringToLevel(viper.GetString("log-level")),
				AddSource: true,
			},
		),
	)
}
