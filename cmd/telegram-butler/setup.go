package main

import (
	"context"
	"fmt"
	"log/slog"

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
	bot, err := telegram.Bot(botCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("bot: %w", err)
	}

	log.Debug("Bot is set")

	if err = commands.Sync(bot); err != nil {
		return nil, nil, fmt.Errorf("sync commands failed: %w", err)
	}

	log.Debug("Commands are synced")

	webhookCfg, err := config.Webhook()
	if err != nil {
		return nil, nil, fmt.Errorf("webhook config: %w", err)
	}

	log.Debug("Webhook is configured", slog.Any("webhook", webhookCfg))

	updates, err := telegram.Webhook(ctx, webhookCfg, bot)
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

	run = func() error {
		go func() {
			if webhookErr := bot.StartWebhook(":" + viper.GetString("port")); webhookErr != nil {
				log.Error("Starting webhook failed", slog.Any("error", webhookErr))
			}
		}()

		bh.Start()
		log.Debug("Handler started")

		return nil
	}

	stop = func() error {
		bh.Stop()
		log.Debug("Handler stopped")

		if err = bot.StopWebhookWithContext(ctx); err != nil {
			return fmt.Errorf("stop webhook: %w", err)
		}

		return nil
	}

	return
}
