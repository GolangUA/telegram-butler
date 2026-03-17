//go:build local

package main

import (
	"context"
	"fmt"
	"log/slog"

	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/viper"
	"golang.ngrok.com/ngrok/v2"
)

func preSetup(ctx context.Context, log *slog.Logger) error {
	log.Info("Setting up ngrok tunnel")

	port := viper.GetString("port")
	fw, err := ngrok.Forward(ctx, ngrok.WithUpstream("http://:"+port))
	if err != nil {
		return fmt.Errorf("start ngrok tunnel: %w", err)
	}
	urlStr := fw.URL().String()
	log.Info("Ngrok tunnel", slog.String("url", urlStr))

	viper.Set("webhook-url", urlStr+"/webhook")

	return nil
}
