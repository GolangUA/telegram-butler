package echo

import (
	"context"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/GolangUA/telegram-butler/internal/module/logger"
)

func Register(bh *th.BotHandler) {
	h := &handler{}
	bh.HandleMessageCtx(h.echo)
}

type handler struct{}

func (h *handler) echo(ctx context.Context, bot *telego.Bot, message telego.Message) {
	log := logger.FromContext(ctx)

	_, err := bot.SendMessage(tu.Message(message.Chat.ChatID(), message.Text))
	if err != nil {
		log.Errorf("Echo message: %s", err)
	}

	log.Debugf("Echo: [%s] %s", message.From.FirstName, message.Text)
}
