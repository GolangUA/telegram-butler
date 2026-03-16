package mute

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/GolangUA/telegram-butler/internal/duration"
	"github.com/GolangUA/telegram-butler/internal/messages"
	"github.com/GolangUA/telegram-butler/internal/module/logger"
	"github.com/GolangUA/telegram-butler/internal/module/telegram"
)

const (
	commandMute      = "m"
	commandMuteFull  = "mute"
	errorDeleteDelay = 15 * time.Second
)

func Register(bh *th.BotHandler) {
	h := &handler{}

	bh.HandleMessage(h.handleMute, th.Or(
		th.CommandEqual(commandMute),
		th.CommandEqual(commandMuteFull),
	))
}

type handler struct{}

func (h *handler) handleMute(ctx *th.Context, message telego.Message) error {
	log := logger.FromContext(ctx)

	log = log.With(slog.Group("caller",
		slog.String("username", message.From.Username),
		slog.Int64("id", message.From.ID),
	))

	log.Log(ctx, logger.LevelTrace, "handling /mute command",
		slog.Int64("chat_id", message.Chat.ID),
		slog.Int("thread_id", message.MessageThreadID),
	)

	chatID := message.Chat.ChatID()

	if !h.isAdmin(ctx, chatID, message.From.ID) {
		return nil
	}

	cmd, err := parseMuteCommand(message.Text)
	if err != nil {
		h.replyWithError(ctx, log, message, err.Error())
		return nil
	}

	target, err := h.resolveTarget(message)
	if err != nil {
		h.replyWithError(ctx, log, message, err.Error())
		return nil
	}

	err = h.verifyNotAdmin(ctx, chatID, target.ID)
	if err != nil {
		h.replyWithError(ctx, log, message, err.Error())
		return nil
	}

	err = h.restrictUser(ctx, chatID, target.ID, cmd.Duration)
	if err != nil {
		log.Error("Failed to restrict user", slog.Any("error", err))

		h.replyWithError(ctx, log, message, "failed to mute user")

		return nil
	}

	targetName := h.displayName(target)

	log.Info("User muted",
		slog.String("target", targetName),
		slog.Int64("target_id", target.ID),
		slog.String("duration", cmd.Duration.String()),
		slog.String("reason", cmd.Reason),
	)

	notifyErr := h.notifyMute(ctx, message, targetName, cmd)
	if notifyErr != nil {
		log.Error("Failed to send mute notification", slog.Any("error", notifyErr))
	}

	return nil
}

func (h *handler) isAdmin(ctx *th.Context, chatID telego.ChatID, userID int64) bool {
	member, err := ctx.Bot().GetChatMember(ctx, &telego.GetChatMemberParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return false
	}

	status := member.MemberStatus()

	return status == telego.MemberStatusCreator || status == telego.MemberStatusAdministrator
}

func (h *handler) verifyNotAdmin(ctx *th.Context, chatID telego.ChatID, userID int64) error {
	member, err := ctx.Bot().GetChatMember(ctx, &telego.GetChatMemberParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to check target user: %w", err)
	}

	status := member.MemberStatus()
	if status == telego.MemberStatusCreator || status == telego.MemberStatusAdministrator {
		return errors.New("cannot mute an admin")
	}

	return nil
}

func (h *handler) restrictUser(
	ctx *th.Context, chatID telego.ChatID, userID int64, duration time.Duration,
) error {
	return ctx.Bot().RestrictChatMember(ctx, &telego.RestrictChatMemberParams{
		ChatID:      chatID,
		UserID:      userID,
		UntilDate:   time.Now().Add(duration).Unix(),
		Permissions: telegram.DenyAllPermissions(),
	})
}

func (h *handler) notifyMute(
	ctx *th.Context, message telego.Message, targetName string, cmd *command,
) error {
	notification := fmt.Sprintf(messages.MuteNotification,
		targetName, message.From.Username, duration.Format(cmd.Duration))

	if cmd.Reason != "" {
		notification += "\nReason: " + cmd.Reason
	}

	_, err := ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:          message.Chat.ChatID(),
		MessageThreadID: message.MessageThreadID,
		ParseMode:       telego.ModeMarkdownV2,
		Text:            notification,
	})

	return err
}

// resolveTarget extracts the target user from the replied message.
// Can be moved to a shared package if reused by other handlers.
func (*handler) resolveTarget(message telego.Message) (*telego.User, error) {
	if message.ReplyToMessage == nil || message.ReplyToMessage.From == nil {
		return nil, errors.New("command must be a reply to the target user's message")
	}

	return message.ReplyToMessage.From, nil
}

func (h *handler) replyWithError(ctx *th.Context, log *slog.Logger, message telego.Message, errText string) {
	err := h.sendAndCleanup(ctx, message, errText)
	if err != nil {
		log.Error("Failed to send error reply", slog.Any("error", err))
	}
}

func (h *handler) sendAndCleanup(ctx *th.Context, message telego.Message, errText string) error {
	reply, err := ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:          message.Chat.ChatID(),
		MessageThreadID: message.MessageThreadID,
		ParseMode:       telego.ModeMarkdownV2,
		Text:            fmt.Sprintf(messages.MuteError, message.Text, errText),
		ReplyParameters: &telego.ReplyParameters{
			MessageID: message.MessageID,
		},
	})
	if err != nil {
		return err
	}

	bot := ctx.Bot()
	chatID := message.Chat.ChatID()
	cmdMessageID := message.MessageID
	replyMessageID := reply.MessageID

	time.AfterFunc(errorDeleteDelay, func() {
		deleteCtx, cancel := context.WithTimeout(context.Background(), errorDeleteDelay)
		defer cancel()

		_ = bot.DeleteMessage(deleteCtx, tu.Delete(chatID, cmdMessageID))
		_ = bot.DeleteMessage(deleteCtx, tu.Delete(chatID, replyMessageID))
	})

	return nil
}

// displayName returns username if available, otherwise first name.
// Can be moved to a shared package if reused by other handlers.
func (*handler) displayName(user *telego.User) string {
	if user.Username != "" {
		return user.Username
	}

	return user.FirstName
}
