package message

import (
	"context"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"

	"github.com/GolangUA/telegram-butler/internal/handler/message/commands"
	"github.com/GolangUA/telegram-butler/internal/module/logger"
)

func Register(bh *th.BotHandler) {
	h := &handler{}
	bh.HandleMessageCtx(h.rules, th.CommandEqual(commands.SendRules))
	bh.HandleMessageCtx(h.usefulInfo, th.CommandEqual(commands.SendUsefulInfo))
}

type handler struct{}

func (h *handler) rules(ctx context.Context, bot *telego.Bot, message telego.Message) {
	log := logger.FromContext(ctx)
	_, err := bot.SendMessage(&telego.SendMessageParams{
		ChatID:    message.Chat.ChatID(),
		ParseMode: telego.ModeHTML,
		Text:      rulesMessage,
	})
	if err != nil {
		log.Errorf("Send rules message error: %s", err)
	}
}

func (h *handler) usefulInfo(ctx context.Context, bot *telego.Bot, message telego.Message) {
	log := logger.FromContext(ctx)
	_, err := bot.SendMessage(&telego.SendMessageParams{
		ChatID:    message.Chat.ChatID(),
		ParseMode: telego.ModeHTML,
		Text:      usefulInfoMessage,
	})
	if err != nil {
		log.Errorf("Send useful info message error: %s", err)
	}
}
