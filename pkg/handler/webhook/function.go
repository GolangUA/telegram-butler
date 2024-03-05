package webhook

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/mymmrac/telego"
	"github.com/spf13/viper"

	_ "github.com/GolangUA/telegram-butler/pkg/config"
)

const FunctionName = "telegram-webhook"

func init() {
	functions.HTTP(FunctionName, telegramWebhook)
}

func telegramWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	secretBytes := sha512.Sum512([]byte(viper.GetString("bot-token")))
	secretToken := hex.EncodeToString(secretBytes[:])

	if r.Header.Get(telego.WebhookSecretTokenHeader) != secretToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var update telego.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("decode update", "error", err)
		return
	}

	if err := handleUpdate(r.Context(), update); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("handle update", "error", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
