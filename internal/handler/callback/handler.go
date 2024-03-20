package callback

import (
	"context"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/GolangUA/telegram-butler/internal/handler/callback/callbackdata"
	"github.com/GolangUA/telegram-butler/internal/module/logger"
)

const BanMessage = "Ваш запит відхилено. У разі помилки зверніться до адміністратора (@vpakh)."

func Register(bh *th.BotHandler) {
	h := &handler{}
	bh.HandleCallbackQueryCtx(h.callbackQuery)
}

type handler struct{}

func (h *handler) callbackQuery(ctx context.Context, bot *telego.Bot, query telego.CallbackQuery) {
	log := logger.FromContext(ctx)
	log.Infof(
		"[CALLBACK QUERY] username: %s, firstname: %s, id: %v",
		query.From.Username,
		query.From.FirstName,
		query.From.ID,
	)

	data, err := callbackdata.Parse(query.Data)
	if err != nil {
		log.Errorf("Parsing callback query data failed: %v", err)
		return
	}

	switch data.Decision {
	case callbackdata.AgreeDecision:
		err := bot.ApproveChatJoinRequest(&telego.ApproveChatJoinRequestParams{
			UserID: query.From.ID,
			ChatID: tu.ID(data.GroupID),
		})
		if err != nil {
			log.Errorf("Join request approve error: %v", err)
		}

	case callbackdata.DeclineDecision:
		err := bot.DeclineChatJoinRequest(&telego.DeclineChatJoinRequestParams{
			UserID: query.From.ID,
			ChatID: tu.ID(data.GroupID),
		})
		if err != nil {
			log.Errorf("Decline join request error: %v", err)
			return
		}

		_, err = bot.SendMessage(tu.Message(tu.ID(query.From.ID), BanMessage))
		if err != nil {
			log.Errorf("Send ban message error: %v", err)
			return
		}
	}
}
