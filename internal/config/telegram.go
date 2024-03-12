package config

import (
	"fmt"
	"net/url"

	"github.com/spf13/viper"

	"github.com/GolangUA/telegram-butler/internal/module/telegram"
)

func Bot() telegram.BotConfig {
	return telegram.BotConfig{
		BotToken: viper.GetString("bot-token"),
	}
}

func Webhook() (telegram.WebhookConfig, error) {
	webhookURL, err := url.Parse(viper.GetString("webhook-url"))
	if err != nil {
		return telegram.WebhookConfig{}, fmt.Errorf("parse webhook url: %w", err)
	}

	return telegram.WebhookConfig{
		BotToken:   viper.GetString("bot-token"),
		WebhookURL: *webhookURL,
	}, nil
}
