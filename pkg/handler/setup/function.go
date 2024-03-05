package setup

import (
	"crypto/sha512"
	"encoding/hex"
	"log/slog"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/mymmrac/telego"
	"github.com/spf13/viper"

	_ "github.com/GolangUA/telegram-butler/pkg/config"
)

const FunctionName = "telegram-setup"

func init() {
	functions.HTTP(FunctionName, telegramSetup)
}

func telegramSetup(w http.ResponseWriter, _ *http.Request) {
	bot, err := telego.NewBot(viper.GetString("bot-token"), telego.WithDiscardLogger())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("create bot", "error", err)
		return
	}

	secretBytes := sha512.Sum512([]byte(viper.GetString("bot-token")))
	secretToken := hex.EncodeToString(secretBytes[:])

	err = bot.SetWebhook(&telego.SetWebhookParams{
		URL:            viper.GetString("webhook-url"),
		AllowedUpdates: []string{telego.MessageUpdates},
		SecretToken:    secretToken,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("set webhook", "error", err)
		return
	}
	slog.Info("successful telegram setup")

	w.WriteHeader(http.StatusOK)
}
