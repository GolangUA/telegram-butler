package callback

import (
	"context"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/GolangUA/telegram-butler/internal/module/logger"
)

const BanMessage = "Ваш запит відхилено. У разі помилки зверніться до адміністратора."

const (
	AgreeDecision   = "agree"
	DeclineDecision = "decline"
)

func Register(bh *th.BotHandler) {
	h := handler{}
	bh.HandleCallbackQueryCtx(h.callbackQuery)
}

type handler struct{}

func (h *handler) callbackQuery(
	ctx context.Context,
	bot *telego.Bot,
	query telego.CallbackQuery,
) {
	log := logger.FromContext(ctx)
	log.Infof(
		"[CALLBACK QUERY] username: %s, firstname: %s, id: %v",
		query.From.Username,
		query.From.FirstName,
		query.From.ID,
	)

	splits := strings.Split(query.Data, "_")
	if len(splits) != 2 {
		log.Errorf("Invalid callback query data token: %v", splits)
		return
	}

	groupID, err := strconv.ParseInt(splits[1], 10, 64)
	if err != nil {
		log.Errorf("Invalid groupID in callback query data: %s", splits[1])
		return
	}

	decision := splits[0]
	switch decision {
	case AgreeDecision:
		err := bot.ApproveChatJoinRequest(&telego.ApproveChatJoinRequestParams{
			UserID: query.From.ID,
			ChatID: tu.ID(groupID),
		})
		if err != nil {
			log.Errorf("Join request approve error: %v", err)
		}

	case DeclineDecision:
		_, err := bot.SendMessage(tu.Message(tu.ID(query.From.ID), BanMessage))
		if err != nil {
			log.Errorf("Send ban message error: %v", err)
		}
		// TODO: ban
	}
}
