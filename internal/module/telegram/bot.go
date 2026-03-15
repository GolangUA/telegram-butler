package telegram

import (
	"context"
	"fmt"

	"github.com/mymmrac/telego"
)

func Bot(ctx context.Context, cfg BotConfig) (*telego.Bot, error) {
	bot, err := telego.NewBot(
		cfg.BotToken,
		telego.WithDiscardLogger(),
		telego.WithHealthCheck(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	if err := syncInfo(ctx, bot); err != nil {
		return nil, fmt.Errorf("sync info failed: %w", err)
	}

	return bot, nil
}

type BotConfig struct {
	BotToken string
}
