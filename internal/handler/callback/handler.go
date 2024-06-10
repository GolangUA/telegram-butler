package callback

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/spf13/viper"

	"github.com/GolangUA/telegram-butler/internal/handler/callback/callbackdata"
	"github.com/GolangUA/telegram-butler/internal/messages"
	"github.com/GolangUA/telegram-butler/internal/module/logger"
)

func Register(bh *th.BotHandler) {
	h := &handler{}
	bh.HandleCallbackQueryCtx(h.callbackQuery)
}

type handler struct{}

func (h *handler) callbackQuery(ctx context.Context, bot *telego.Bot, query telego.CallbackQuery) {
	log := logger.FromContext(ctx)

	log = log.With(slog.Group("user",
		slog.String("username", query.From.Username),
		slog.Int64("id", query.From.ID),
	))

	log.Info("[CALLBACK QUERY]")

	if err := bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID)); err != nil {
		log.Error("Sending answer to callback query failed", slog.Any("error", err))
	}

	data, err := callbackdata.Parse(query.Data)
	if err != nil {
		log.Error("Parsing callback query data failed", slog.Any("error", err))
		return
	}

	var msg string
	switch data.Decision {
	case callbackdata.AgreeDecision:
		err = bot.ApproveChatJoinRequest(&telego.ApproveChatJoinRequestParams{
			UserID: query.From.ID,
			ChatID: tu.ID(data.GroupID),
		})
		if err != nil {
			log.Error("Join request approve error", slog.Any("error", err))
		}

		log.Info("Successfully approved join request")
		msg = fmt.Sprintf(messages.Welcome, query.From.FirstName, viper.GetString("group-name"))
	case callbackdata.DeclineDecision:
		err = bot.DeclineChatJoinRequest(&telego.DeclineChatJoinRequestParams{
			UserID: query.From.ID,
			ChatID: tu.ID(data.GroupID),
		})
		if err != nil {
			log.Error("Decline join request failed", slog.Any("error", err))
			return
		}

		msg = fmt.Sprintf(messages.Decline, viper.GetString("admin-username"))
	}

	_, err = bot.EditMessageText(&telego.EditMessageTextParams{
		MessageID: data.MessageID,
		ChatID:    tu.ID(query.From.ID),
		Text:      msg,
	})
	if err != nil {
		log.Error("Sending decision message failed", slog.Any("error", err))
	}
}
