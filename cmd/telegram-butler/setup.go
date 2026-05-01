package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/spf13/viper"

	"github.com/GolangUA/telegram-butler/internal/config"
	"github.com/GolangUA/telegram-butler/internal/handler/callback"
	"github.com/GolangUA/telegram-butler/internal/handler/join"
	"github.com/GolangUA/telegram-butler/internal/handler/message"
	"github.com/GolangUA/telegram-butler/internal/handler/message/commands"
	"github.com/GolangUA/telegram-butler/internal/handler/mute"
	"github.com/GolangUA/telegram-butler/internal/handler/report"
	"github.com/GolangUA/telegram-butler/internal/module/telegram"
	"github.com/GolangUA/telegram-butler/internal/repository/memory"
	reportsvc "github.com/GolangUA/telegram-butler/internal/service/report"
)

func setup(ctx context.Context, log *slog.Logger) (run func() error, stop func() error, err error) {
	log.Info("Setting up the Bot")

	botCfg := config.Bot()

	bot, err := telegram.Bot(ctx, botCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("bot: %w", err)
	}

	log.Debug("Bot is set")

	err = commands.Sync(ctx, bot)
	if err != nil {
		return nil, nil, fmt.Errorf("sync commands failed: %w", err)
	}

	log.Debug("Commands are synced")

	webhookCfg, err := config.Webhook()
	if err != nil {
		return nil, nil, fmt.Errorf("webhook config: %w", err)
	}

	log.Debug("Webhook is configured", slog.Any("webhook", webhookCfg))

	mux := http.NewServeMux()

	updates, err := telegram.Webhook(ctx, webhookCfg, bot, mux)
	if err != nil {
		return nil, nil, fmt.Errorf("webhook: %w", err)
	}

	log.Debug("Updates channel is configured")

	bh, err := telegram.NewHandler(ctx, bot, updates)
	if err != nil {
		return nil, nil, fmt.Errorf("bot handler: %w", err)
	}

	voteRepo := memory.NewVoteRepository()
	voteSvc := reportsvc.NewService(voteRepo, bot)

	message.Register(bh)
	join.Register(bh)
	callback.Register(bh)
	mute.Register(bh)
	report.Register(bh, bot, voteSvc)

	log.Debug("Bot handlers are registered")

	srv := &http.Server{
		Addr:              ":" + viper.GetString("port"),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	run = func() error {
		go func() {
			webhookErr := srv.ListenAndServe()
			if webhookErr != nil && webhookErr != http.ErrServerClosed {
				log.Error("Starting webhook failed", slog.Any("error", webhookErr))
			}
		}()

		startErr := bh.Start()
		if startErr != nil {
			return fmt.Errorf("bot handler start: %w", startErr)
		}

		log.Debug("Handler stopped processing updates")

		return nil
	}

	stop = func() error {
		stopErr := bh.Stop()
		if stopErr != nil {
			log.Error("Bot handler stop failed", slog.Any("error", stopErr))
		}

		log.Debug("Handler stopped")

		err = srv.Shutdown(ctx)
		if err != nil {
			return fmt.Errorf("stop webhook: %w", err)
		}

		return nil
	}

	return
}
