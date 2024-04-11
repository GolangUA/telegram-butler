package callback

import (
	"context"
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/spf13/viper"

	"github.com/GolangUA/telegram-butler/internal/handler/callback/callbackdata"
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

	switch data.Decision {
	case callbackdata.AgreeDecision:
		err := bot.ApproveChatJoinRequest(&telego.ApproveChatJoinRequestParams{
			UserID: query.From.ID,
			ChatID: tu.ID(data.GroupID),
		})
		if err != nil {
			log.Error("Join request approve error", slog.Any("error", err))
		}

		msg := getWelcomeMessage(query.From.FirstName, viper.GetString("group-name"))
		_, err = bot.SendMessage(&telego.SendMessageParams{
			ChatID:    tu.ID(query.From.ID),
			ParseMode: telego.ModeHTML,
			Text:      msg,
		})
		if err != nil {
			log.Error("Sending welcome message failed", slog.Any("error", err))
		}

	case callbackdata.DeclineDecision:
		err := bot.DeclineChatJoinRequest(&telego.DeclineChatJoinRequestParams{
			UserID: query.From.ID,
			ChatID: tu.ID(data.GroupID),
		})
		if err != nil {
			log.Error("Decline join request failed", slog.Any("error", err))
			return
		}

		msg := getBanMessage(viper.GetString("admin-username"))
		_, err = bot.SendMessage(tu.Message(tu.ID(query.From.ID), msg))
		if err != nil {
			log.Error("Sending ban message failed", slog.Any("error", err))
		}
	}
}
