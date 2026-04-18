package telegram

import (
	"context"
	"fmt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"

	"github.com/GolangUA/telegram-butler/internal/module/logger"
)

func NewHandler(ctx context.Context, bot *telego.Bot, updates <-chan telego.Update) (*th.BotHandler, error) {
	log := logger.FromContext(ctx)

	bh, err := th.NewBotHandler(bot, updates)
	if err != nil {
		return nil, fmt.Errorf("create bot handler: %w", err)
	}

	// Inject logger into handler context so handlers can use logger.FromContext(ctx)
	bh.Use(func(thCtx *th.Context, update telego.Update) error {
		thCtx = thCtx.WithContext(logger.ToContext(thCtx, log))
		return thCtx.Next(update)
	})

	return bh, nil
}
