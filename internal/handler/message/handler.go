package message

import (
	"context"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"

	"github.com/GolangUA/telegram-butler/internal/module/logger"
)

func Register(bh *th.BotHandler) {
	h := &handler{}
	bh.HandleMessageCtx(h.message)
}

type handler struct{}

func (h *handler) message(ctx context.Context, bot *telego.Bot, message telego.Message) {
	log := logger.FromContext(ctx)

	if message.Text == sendRulesCommand {
		_, err := bot.SendMessage(&telego.SendMessageParams{
			ChatID:    message.Chat.ChatID(),
			ParseMode: telego.ModeHTML,
			Text:      rulesMessage,
		})
		if err != nil {
			log.Errorf("Send rules message error: %s", err)
		}
	}
}
