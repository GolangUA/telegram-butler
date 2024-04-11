//go:build !local

package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/viper"

	"github.com/GolangUA/telegram-butler/internal/module/gcp/cloudrun"
	"github.com/GolangUA/telegram-butler/internal/module/gcp/secrets"
)

func preSetup(ctx context.Context, log *slog.Logger) error {
	log.Info("Initializing application on GCP")

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

	log.Info("Finished setting up", slog.String("url", cloudRunURL))

	viper.Set("webhook-url", cloudRunURL+"/webhook")

	return nil
}
