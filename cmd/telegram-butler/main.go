package main

import (
	"context"
	"os"
	"os/signal"

	_ "github.com/GolangUA/telegram-butler/pkg/config"
	"github.com/GolangUA/telegram-butler/pkg/module/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	log := logger.NewSTDLogger()
	ctx = logger.ToContext(ctx, log)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	log.Infof("Starting...")
	err := preSetup(ctx, log)
	if err != nil {
		log.Fatalf("Pre-setup: %s", err)
	}

	run, stop, err := setup(ctx, log)
	if err != nil {
		log.Fatalf("Setup: %s", err)
	}
	go func() {
		if runErr := run(); runErr != nil {
			log.Errorf("Run: %s", runErr)
		}
		sigs <- os.Interrupt
	}()

	done := make(chan struct{}, 1)
	go func() {
		<-sigs
		log.Infof("Stopping...")
		if stopErr := stop(); stopErr != nil {
			log.Errorf("Stop: %s", stopErr)
		}
		cancel()
		done <- struct{}{}
	}()

	<-done
	log.Infof("Bye!")
}
