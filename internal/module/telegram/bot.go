package telegram

import (
	"context"
	"fmt"

	"github.com/mymmrac/telego"
)

func Bot(ctx context.Context, cfg Config) (*telego.Bot, error) {
	bot, err := telego.NewBot(
		cfg.Token,
		telego.WithDiscardLogger(),
		telego.WithHealthCheck(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	err = syncInfo(ctx, bot)
	if err != nil {
		return nil, fmt.Errorf("sync info failed: %w", err)
	}

	return bot, nil
}

type Config struct {
	Token string
}
