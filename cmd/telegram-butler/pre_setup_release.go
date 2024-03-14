//go:build !local

package main

import (
	"context"
	"fmt"

	"github.com/spf13/viper"

	"github.com/GolangUA/telegram-butler/internal/module/gcp/cloudrun"
	"github.com/GolangUA/telegram-butler/internal/module/gcp/secrets"
	"github.com/GolangUA/telegram-butler/internal/module/logger"
)

func preSetup(ctx context.Context, _ logger.Logger) error {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	secretManager, err := secrets.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("initialize secret manager client: %w", err)
	}

	botTokenSecretName := secrets.BuildSecretName(viper.GetString("project-id"), secrets.BotTokenSecretID, "latest")
	botToken, err := secretManager.GetSecretValue(ctx, botTokenSecretName)
	if err != nil {
		return fmt.Errorf("get bot token: %w", err)
	}

	viper.Set("bot-token", botToken)

	cloudRunURL, err := cloudrun.GetServiceURL(ctx)
	if err != nil {
		return fmt.Errorf("get cloud run URL: %w", err)
	}

	viper.Set("webhook-url", cloudRunURL+"/webhook")

	return nil
}
