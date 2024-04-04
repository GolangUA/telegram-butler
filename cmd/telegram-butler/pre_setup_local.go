//go:build local

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/viper"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
)

func preSetup(ctx context.Context, log *slog.Logger) error {
	log.Info("Setting up ngrok tunnel")

	fw, err := ngrok.ListenAndForward(
		ctx,
		&url.URL{
			Scheme: "http",
			Host:   ":" + viper.GetString("port"),
		},
		config.HTTPEndpoint(),
		ngrok.WithAuthtokenFromEnv(),
	)
	if err != nil {
		return fmt.Errorf("start ngrok tunnel: %w", err)
	}
	log.Info("Ngrok tunnel", slog.String("url", fw.URL()))

	viper.Set("webhook-url", fw.URL()+"/webhook")

	return nil
}
