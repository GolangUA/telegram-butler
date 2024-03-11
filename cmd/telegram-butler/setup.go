package main

import (
	"context"
	"fmt"

	"github.com/spf13/viper"

	"github.com/GolangUA/telegram-butler/pkg/config"
	"github.com/GolangUA/telegram-butler/pkg/handler/echo"
	"github.com/GolangUA/telegram-butler/pkg/module/logger"
	"github.com/GolangUA/telegram-butler/pkg/module/telegram"
)

func setup(ctx context.Context, log logger.Logger) (run func() error, stop func() error, err error) {
	botCfg := config.Bot()
	bot, err := telegram.Bot(botCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("bot: %w", err)
	}

	webhookCfg, err := config.Webhook()
	if err != nil {
		return nil, nil, fmt.Errorf("webhook config: %w", err)
	}

	updates, err := telegram.Webhook(ctx, webhookCfg, bot)
	if err != nil {
		return nil, nil, fmt.Errorf("webhook: %w", err)
	}

	bh, err := telegram.BotHandler(ctx, bot, updates)
	if err != nil {
		return nil, nil, fmt.Errorf("bot handler: %w", err)
	}

	echo.Register(bh)

	run = func() error {
		go func() {
			if webhookErr := bot.StartWebhook(":" + viper.GetString("port")); webhookErr != nil {
				log.Errorf("Start webhook: %s", webhookErr)
			}
		}()

		bh.Start()
		return nil
	}

	stop = func() error {
		bh.Stop()

		if err = bot.StopWebhookWithContext(ctx); err != nil {
			return fmt.Errorf("stop webhook: %w", err)
		}

		return nil
	}

	return
}
