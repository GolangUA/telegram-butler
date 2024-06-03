package join

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/spf13/viper"

	"github.com/GolangUA/telegram-butler/internal/handler/callback/callbackdata"
	"github.com/GolangUA/telegram-butler/internal/handler/join/validator"
	"github.com/GolangUA/telegram-butler/internal/messages"
	"github.com/GolangUA/telegram-butler/internal/module/logger"
)

const (
	AgreeText     = "Згоден"
	DontAgreeText = "Не згоден"
)

func Register(bh *th.BotHandler) {
	h := &handler{
		validator: validator.New(validator.DefaultForbiddenList),
	}
	bh.HandleChatJoinRequestCtx(h.chatJoinRequest)
}

type nameValidator interface {
	Validate(names ...string) bool
}

type handler struct {
	validator nameValidator
}

func (h *handler) chatJoinRequest(ctx context.Context, bot *telego.Bot, request telego.ChatJoinRequest) {
	log := logger.FromContext(ctx)

	log = log.With(slog.Group("user",
		slog.String("username", request.From.Username),
		slog.Int64("id", request.From.ID),
	))

	log.Info("[JOIN REQUEST]")

	if !h.validator.Validate(request.From.FirstName, request.From.LastName, request.From.Username) {
		log.Info("Name validation is failed")
		err := bot.DeclineChatJoinRequest(&telego.DeclineChatJoinRequestParams{
			UserID: request.From.ID,
			ChatID: tu.ID(request.Chat.ID),
		})
		if err != nil {
			log.Error("Decline join request failed", slog.Any("error", err))
			return
		}

		msg := fmt.Sprintf(messages.Decline, viper.GetString("admin-username"))
		_, err = bot.SendMessage(&telego.SendMessageParams{
			ChatID:      tu.ID(request.From.ID),
			Text:        msg,
			ReplyMarkup: tu.ReplyKeyboardRemove(),
		})
		if err != nil {
			log.Error("Sending decision message failed", slog.Any("error", err))
		}

		return
	}

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

	_, err := bot.SendMessage(&telego.SendMessageParams{
		ChatID:    tu.ID(request.From.ID),
		ParseMode: telego.ModeHTML,
		Text:      messages.JoinHeader + messages.Rules,
	})
	if err != nil {
		log.Error("Sending TermsOfUse and Rules message failed", slog.Any("error", err))
	}

	msg := tu.Message(tu.ID(request.From.ID), messages.JoinFooter).WithReplyMarkup(k).WithProtectContent()
	if _, err := bot.SendMessage(msg); err != nil {
		log.Error("Sending terms of use failed", slog.Any("error", err))
	}
}
