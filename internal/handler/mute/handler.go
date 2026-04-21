package mute

import (
	"context"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"slices"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/GolangUA/telegram-butler/internal/duration"
	"github.com/GolangUA/telegram-butler/internal/handler/message/commands"
	"github.com/GolangUA/telegram-butler/internal/messages"
	"github.com/GolangUA/telegram-butler/internal/module/logger"
	"github.com/GolangUA/telegram-butler/internal/module/telegram"
)

const (
	errorMessageLifetime = 15 * time.Second
	deleteTimeout        = 5 * time.Second
)

func Register(bh *th.BotHandler) {
	h := &handler{}

	bh.HandleMessage(h.handleMute, th.Or(
		th.CommandEqual(commands.Mute),
		th.CommandEqual(commands.MuteFull),
	))
}

type handler struct{}

func (h *handler) handleMute(ctx *th.Context, message telego.Message) error {
	log := logger.FromContext(ctx)

	log = log.With(slog.Group("user",
		slog.String("username", message.From.Username),
		slog.Int64("id", message.From.ID),
	))

	log.Log(ctx, logger.LevelTrace, "handling /mute command",
		slog.Int64("chat_id", message.Chat.ID),
		slog.Int("thread_id", message.MessageThreadID),
	)

	chatID := message.Chat.ChatID()

	callerIsAdmin, err := h.isAdmin(ctx, chatID, message.From.ID)
	if err != nil {
		log.Error("Failed to check caller admin status", slog.Any("error", err))
		return nil
	}

	if !callerIsAdmin {
		return nil
	}

	cmd, err := parseCommand(message.Text)
	if err != nil {
		h.replyWithError(ctx, log, message, err.Error())
		return nil
	}

	target, err := h.resolveTarget(message)
	if err != nil {
		h.replyWithError(ctx, log, message, err.Error())
		return nil
	}

	targetIsAdmin, err := h.isAdmin(ctx, chatID, target.ID)
	if err != nil {
		log.Error("Failed to check target user status", slog.Any("error", err))

		h.replyWithError(ctx, log, message, "failed to check target user")

		return nil
	}

	if targetIsAdmin {
		h.replyWithError(ctx, log, message, "cannot mute an admin")
		return nil
	}

	err = h.restrictUser(ctx, chatID, target.ID, cmd.Duration)
	if err != nil {
		log.Error("Failed to restrict user", slog.Any("error", err))

		h.replyWithError(ctx, log, message, "failed to mute user")

		return nil
	}

	log.Info("User muted",
		slog.Group("target",
			slog.String("username", target.Username),
			slog.String("first_name", target.FirstName),
			slog.Int64("id", target.ID),
		),
		slog.String("duration", cmd.Duration.String()),
		slog.String("reason", cmd.Reason),
	)

	err = h.notifyMute(ctx, message, target, cmd)
	if err != nil {
		log.Error("Failed to send mute notification", slog.Any("error", err))
	}

	return nil
}

var adminStatuses = []string{
	telego.MemberStatusCreator,
	telego.MemberStatusAdministrator,
}

func (h *handler) isAdmin(ctx *th.Context, chatID telego.ChatID, userID int64) (bool, error) {
	member, err := ctx.Bot().GetChatMember(ctx, &telego.GetChatMemberParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check user status: %w", err)
	}

	return slices.Contains(adminStatuses, member.MemberStatus()), nil
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
	ctx *th.Context, message telego.Message, target *telego.User, cmd *command,
) error {
	notification := fmt.Sprintf(messages.MuteNotification,
		h.mentionUser(target), h.mentionUser(message.From), duration.Format(cmd.Duration))

	if cmd.Reason != "" {
		notification += "\n" + messages.MuteReason + ": " + html.EscapeString(cmd.Reason)
	}

	_, err := ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:          message.Chat.ChatID(),
		MessageThreadID: message.MessageThreadID,
		ParseMode:       telego.ModeHTML,
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
		ParseMode:       telego.ModeHTML,
		Text:            fmt.Sprintf(messages.MuteError, html.EscapeString(message.Text), html.EscapeString(errText)),
		ReplyParameters: &telego.ReplyParameters{
			MessageID: message.MessageID,
		},
	})
	if err != nil {
		return err
	}

	// Capture values for the deferred goroutine — ctx won't be valid after handler returns
	bot := ctx.Bot()
	chatID := message.Chat.ChatID()
	cmdMessageID := message.MessageID
	replyMessageID := reply.MessageID

	time.AfterFunc(errorMessageLifetime, func() {
		deleteCtx, cancel := context.WithTimeout(context.Background(), deleteTimeout)
		defer cancel()

		// Best-effort cleanup — no logger available in deferred goroutine
		_ = bot.DeleteMessage(deleteCtx, tu.Delete(chatID, cmdMessageID))
		_ = bot.DeleteMessage(deleteCtx, tu.Delete(chatID, replyMessageID))
	})

	return nil
}

// mentionUser returns an HTML inline mention link for the user.
// Can be moved to a shared package if reused by other handlers.
func (*handler) mentionUser(user *telego.User) string {
	name := html.EscapeString(user.FirstName)
	if user.Username != "" {
		name = html.EscapeString(user.FirstName) + " (@" + html.EscapeString(user.Username) + ")"
	}

	return fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>", user.ID, name)
}
