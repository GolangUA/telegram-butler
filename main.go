package main

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/viper"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
	slogadapter "golang.ngrok.com/ngrok/log/slog"

	_ "github.com/GolangUA/telegram-butler/pkg/handler/setup"
	"github.com/GolangUA/telegram-butler/pkg/handler/webhook"
)

func main() {
	ctx := context.Background()

	viper.SetDefault("hostname", "localhost")
	viper.SetDefault("port", "8080")

	viper.AutomaticEnv()

	fw, err := ngrok.ListenAndForward(
		ctx,
		&url.URL{
			Scheme: "http",
			Host:   viper.GetString("hostname") + ":" + viper.GetString("port"),
		},
		config.HTTPEndpoint(),
		ngrok.WithAuthtokenFromEnv(),
		ngrok.WithLogger(slogadapter.NewLogger(slog.Default())),
	)
	if err != nil {
		slog.Error("start ngrok tunnel", "error", err)
		return
	}
	slog.Info("ngrok tunnel", "url", fw.URL())

	viper.Set("webhook-url", fw.URL()+"/"+webhook.FunctionName)

	if err = funcframework.StartHostPort(viper.GetString("hostname"), viper.GetString("port")); err != nil {
		slog.Error("start function", "error", err)
		return
	}
}
