package callback

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/spf13/viper"

	"github.com/GolangUA/telegram-butler/internal/handler/callback/callbackdata"
	"github.com/GolangUA/telegram-butler/internal/messages"
	"github.com/GolangUA/telegram-butler/internal/module/logger"
	"github.com/GolangUA/telegram-butler/internal/module/telegram"
)

func Register(bh *th.BotHandler) {
	h := &handler{}

	group := bh.Group()
	group.HandleCallbackQuery(h.handleAgree, th.CallbackDataPrefix(callbackdata.AgreeDecision+"_"))
	group.HandleCallbackQuery(h.handleDecline, th.CallbackDataPrefix(callbackdata.DeclineDecision+"_"))
}

type handler struct{}

func (h *handler) handleAgree(ctx *th.Context, query telego.CallbackQuery) error {
	log := logCtx(ctx, query)
	log.Info("[CALLBACK QUERY] agree")

	answer(ctx, log, query.ID)

	data, err := callbackdata.Parse(query.Data)
	if err != nil {
		log.Error("Parsing callback query data failed", slog.Any("error", err))
		return nil
	}

	err = ctx.Bot().ApproveChatJoinRequest(ctx, &telego.ApproveChatJoinRequestParams{
		UserID: query.From.ID,
		ChatID: tu.ID(data.GroupID),
	})
	if err != nil {
		log.Error("Join request approve error", slog.Any("error", err))
		return nil
	}

	log.Info("Successfully approved join request")

	// 24h read-only for new members
	restrictErr := ctx.Bot().RestrictChatMember(ctx, &telego.RestrictChatMemberParams{
		ChatID:      tu.ID(data.GroupID),
		UserID:      query.From.ID,
		UntilDate:   time.Now().Add(24 * time.Hour).Unix(),
		Permissions: telegram.DenyAllPermissions(),
	})
	if restrictErr != nil {
		log.Error("Failed to restrict new member", slog.Any("error", restrictErr))
	}

	msg := fmt.Sprintf(messages.Welcome, query.From.FirstName, viper.GetString("group-name"))
	editMessage(ctx, log, query.From.ID, data.MessageID, msg)
	return nil
}

func (h *handler) handleDecline(ctx *th.Context, query telego.CallbackQuery) error {
	log := logCtx(ctx, query)
	log.Info("[CALLBACK QUERY] decline")

	answer(ctx, log, query.ID)

	data, err := callbackdata.Parse(query.Data)
	if err != nil {
		log.Error("Parsing callback query data failed", slog.Any("error", err))
		return nil
	}

	err = ctx.Bot().DeclineChatJoinRequest(ctx, &telego.DeclineChatJoinRequestParams{
		UserID: query.From.ID,
		ChatID: tu.ID(data.GroupID),
	})
	if err != nil {
		log.Error("Decline join request failed", slog.Any("error", err))
		return nil
	}

	msg := fmt.Sprintf(messages.Decline, viper.GetString("admin-username"))
	editMessage(ctx, log, query.From.ID, data.MessageID, msg)
	return nil
}

func logCtx(ctx *th.Context, query telego.CallbackQuery) *slog.Logger {
	log := logger.FromContext(ctx).With(slog.Group("user",
		slog.String("username", query.From.Username),
		slog.Int64("id", query.From.ID),
	))

	log.Log(ctx, logger.LevelTrace, "processing callback",
		slog.String("data", query.Data),
	)
	return log
}

func answer(ctx *th.Context, log *slog.Logger, queryID string) {
	err := ctx.Bot().AnswerCallbackQuery(ctx, tu.CallbackQuery(queryID))
	if err != nil {
		log.Error("Sending answer to callback query failed", slog.Any("error", err))
	}
}

func editMessage(ctx *th.Context, log *slog.Logger, userID int64, messageID int, text string) {
	_, err := ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
		MessageID: messageID,
		ChatID:    tu.ID(userID),
		Text:      text,
	})
	if err != nil {
		log.Error("Sending decision message failed", slog.Any("error", err))
	}
}
