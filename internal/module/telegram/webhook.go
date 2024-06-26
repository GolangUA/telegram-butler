package telegram

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/mymmrac/telego"
)

func Webhook(ctx context.Context, cfg WebhookConfig, bot *telego.Bot) (<-chan telego.Update, error) {
	secretBytes := sha512.Sum512([]byte(cfg.BotToken))
	secretToken := hex.EncodeToString(secretBytes[:])

	updates, err := bot.UpdatesViaWebhook(
		cfg.WebhookURL.Path,
		telego.WithWebhookSet(&telego.SetWebhookParams{
			URL: cfg.WebhookURL.String(),
			AllowedUpdates: []string{
				telego.MessageUpdates,
				telego.ChatJoinRequestUpdates,
				telego.CallbackQueryUpdates,
			},
			SecretToken: secretToken,
		}),
		telego.WithWebhookContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("updates via webhook: %w", err)
	}

	return updates, nil
}

type WebhookConfig struct {
	BotToken   string
	WebhookURL url.URL
}

// LogValue satisfies the slog.LogValuer interface for WebhookConfig
func (w WebhookConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("bot_token", "[REDACTED]"),
		slog.String("webhook_host", w.WebhookURL.Host),
	)
}
