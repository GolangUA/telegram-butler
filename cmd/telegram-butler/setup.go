package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/spf13/viper"

	"github.com/GolangUA/telegram-butler/internal/config"
	"github.com/GolangUA/telegram-butler/internal/handler/callback"
	"github.com/GolangUA/telegram-butler/internal/handler/join"
	"github.com/GolangUA/telegram-butler/internal/handler/message"
	"github.com/GolangUA/telegram-butler/internal/handler/message/commands"
	"github.com/GolangUA/telegram-butler/internal/module/telegram"
)

func setup(ctx context.Context, log *slog.Logger) (run func() error, stop func() error, err error) {
	log.Info("Setting up the Bot")

	botCfg := config.Bot()
	bot, err := telegram.Bot(ctx, botCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("bot: %w", err)
	}

	log.Debug("Bot is set")

	if err = commands.Sync(ctx, bot); err != nil {
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

	bh, err := telegram.BotHandler(ctx, bot, updates)
	if err != nil {
		return nil, nil, fmt.Errorf("bot handler: %w", err)
	}

	message.Register(bh)
	join.Register(bh)
	callback.Register(bh)

	log.Debug("Bot handlers are registered")

	srv := &http.Server{
		Addr:    ":" + viper.GetString("port"),
		Handler: mux,
	}

	run = func() error {
		go func() {
			if webhookErr := srv.ListenAndServe(); webhookErr != nil && webhookErr != http.ErrServerClosed {
				log.Error("Starting webhook failed", slog.Any("error", webhookErr))
			}
		}()

		bh.Start()
		log.Debug("Handler stopped processing updates")

		return nil
	}

	stop = func() error {
		bh.Stop()
		log.Debug("Handler stopped")

		if err = srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("stop webhook: %w", err)
		}

		return nil
	}

	return
}
