package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/spf13/viper"

	_ "github.com/GolangUA/telegram-butler/internal/config"
	"github.com/GolangUA/telegram-butler/internal/module/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	log := logger.SetupLogger(
		viper.GetString("log-level"),
		viper.GetString("log-format"),
		viper.GetBool("log-source"),
	)

	ctx = logger.ToContext(ctx, log)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	log.Info("Starting Bot...")
	err := preSetup(ctx, log)
	if err != nil {
		log.Error("Pre-setup", slog.Any("error", err))
		os.Exit(1)
	}

	run, stop, err := setup(ctx, log)
	if err != nil {
		log.Error("Setup", slog.Any("error", err))
		os.Exit(1)
	}

	log.Debug("Starting handling queries")
	go func() {
		if runErr := run(); runErr != nil {
			log.Error("Run", slog.Any("error", runErr))
		}
		sigs <- os.Interrupt
	}()

	done := make(chan struct{}, 1)
	go func() {
		<-sigs
		log.Info("Stopping Bot handler...")
		if stopErr := stop(); stopErr != nil {
			log.Error("Stop", slog.Any("error", stopErr))
		}
		cancel()
		done <- struct{}{}
	}()

	<-done
	log.Info("Finished...")
}
