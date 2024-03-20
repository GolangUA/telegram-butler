package join

import (
	"context"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/GolangUA/telegram-butler/internal/handler/callback/callbackdata"
	"github.com/GolangUA/telegram-butler/internal/module/logger"
)

const TermsOfUse = `
Приймаючи запрошення у цю групу ви автоматично:
	1. Засуджуєте війну рф проти України.
	2. Не визнаєте тимчасовано окупованії українські території субʼєктом рф.
	3. Приймаєте правила спільноти (правила групи: /rules).
`

const (
	AgreeText     = "Згоден"
	DontAgreeText = "Не згоден"
)

func Register(bh *th.BotHandler) {
	h := &handler{}
	bh.HandleChatJoinRequestCtx(h.chatJoinRequest)
}

type handler struct{}

func (h *handler) chatJoinRequest(ctx context.Context, bot *telego.Bot, request telego.ChatJoinRequest) {
	log := logger.FromContext(ctx)
	log.Infof(
		"[JOIN REQUEST] username: %s, firstname: %s, id: %v",
		request.From.Username,
		request.From.FirstName,
		request.From.ID,
	)

	tu.Keyboard()
	k := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			telego.InlineKeyboardButton{
				Text:         AgreeText,
				CallbackData: callbackdata.NewAgreeWithGroupID(request.Chat.ID),
			},
			telego.InlineKeyboardButton{
				Text:         DontAgreeText,
				CallbackData: callbackdata.NewDeclineWithGroupID(request.Chat.ID),
			},
		),
	)

	msg := tu.Message(tu.ID(request.From.ID), TermsOfUse).WithReplyMarkup(k).WithProtectContent()
	if _, err := bot.SendMessage(msg); err != nil {
		log.Errorf("Send terms of use error: %v", err)
	}
}
