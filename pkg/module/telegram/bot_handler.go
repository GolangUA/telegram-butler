package telegram

import (
	"context"
	"fmt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func BotHandler(ctx context.Context, bot *telego.Bot, updates <-chan telego.Update) (*th.BotHandler, error) {
	bh, err := th.NewBotHandler(
		bot,
		updates,
		th.WithDone(ctx.Done()),
	)
	if err != nil {
		return nil, fmt.Errorf("create bot handler: %w", err)
	}

	// Override update context
	bh.Use(func(bot *telego.Bot, update telego.Update, next th.Handler) {
		next(bot, update.WithContext(ctx))
	})

	return bh, nil
}
