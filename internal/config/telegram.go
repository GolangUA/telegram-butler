package config

import (
	"fmt"
	"net/url"

	"github.com/spf13/viper"

	"github.com/GolangUA/telegram-butler/internal/module/telegram"
)

func Bot() telegram.Config {
	return telegram.Config{
		Token: viper.GetString("bot-token"),
	}
}

func Webhook() (telegram.WebhookConfig, error) {
	webhookURL, err := url.Parse(viper.GetString("webhook-url"))
	if err != nil {
		return telegram.WebhookConfig{}, fmt.Errorf("parse webhook url: %w", err)
	}

	return telegram.WebhookConfig{
		Token: viper.GetString("bot-token"),
		URL:   *webhookURL,
	}, nil
}
