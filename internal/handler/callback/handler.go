package callback

import (
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
	bh.HandleCallbackQuery(h.callbackQuery)
}

type handler struct{}

func (h *handler) callbackQuery(ctx *th.Context, query telego.CallbackQuery) error {
	log := logger.FromContext(ctx)

	log = log.With(slog.Group("user",
		slog.String("username", query.From.Username),
		slog.Int64("id", query.From.ID),
	))

	log.Info("[CALLBACK QUERY]")

	log.Log(ctx, logger.LevelTrace, "processing callback",
		slog.String("data", query.Data),
	)

	if err := ctx.Bot().AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID)); err != nil {
		log.Error("Sending answer to callback query failed", slog.Any("error", err))
	}

	data, err := callbackdata.Parse(query.Data)
	if err != nil {
		log.Error("Parsing callback query data failed", slog.Any("error", err))
		return nil
	}

	var msg string
	switch data.Decision {
	case callbackdata.AgreeDecision:
		err = ctx.Bot().ApproveChatJoinRequest(ctx, &telego.ApproveChatJoinRequestParams{
			UserID: query.From.ID,
			ChatID: tu.ID(data.GroupID),
		})
		if err != nil {
			log.Error("Join request approve error", slog.Any("error", err))
		}

		log.Info("Successfully approved join request")
		msg = fmt.Sprintf(messages.Welcome, query.From.FirstName, viper.GetString("group-name"))
	case callbackdata.DeclineDecision:
		err = ctx.Bot().DeclineChatJoinRequest(ctx, &telego.DeclineChatJoinRequestParams{
			UserID: query.From.ID,
			ChatID: tu.ID(data.GroupID),
		})
		if err != nil {
			log.Error("Decline join request failed", slog.Any("error", err))
			return nil
		}

		msg = fmt.Sprintf(messages.Decline, viper.GetString("admin-username"))
	}

	_, err = ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
		MessageID: data.MessageID,
		ChatID:    tu.ID(query.From.ID),
		Text:      msg,
	})
	if err != nil {
		log.Error("Sending decision message failed", slog.Any("error", err))
	}

	return nil
}
