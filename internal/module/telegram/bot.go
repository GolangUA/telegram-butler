package telegram

import (
	"fmt"

	"github.com/mymmrac/telego"
)

func Bot(cfg BotConfig) (*telego.Bot, error) {
	bot, err := telego.NewBot(
		cfg.BotToken,
		telego.WithDiscardLogger(),
		telego.WithHealthCheck(),
	)
	if err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	return bot, nil
}

type BotConfig struct {
	BotToken string
}
