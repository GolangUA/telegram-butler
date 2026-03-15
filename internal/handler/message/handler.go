package message

import (
	"fmt"
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"github.com/spf13/viper"

	"github.com/GolangUA/telegram-butler/internal/handler/message/commands"
	"github.com/GolangUA/telegram-butler/internal/messages"
	"github.com/GolangUA/telegram-butler/internal/module/logger"
)

func Register(bh *th.BotHandler) {
	h := &handler{}
	bh.HandleMessage(h.rules, th.CommandEqual(commands.SendRules))
	bh.HandleMessage(h.usefulInfo, th.CommandEqual(commands.SendUsefulInfo))
	bh.HandleMessage(h.help, th.CommandEqual(commands.SendHelp))
}

type handler struct{}

func (h *handler) rules(ctx *th.Context, message telego.Message) error {
	log := logger.FromContext(ctx)

	log = log.With(slog.Group("user",
		slog.String("username", message.From.Username),
		slog.Int64("id", message.From.ID),
	))

	log.Log(ctx, logger.LevelTrace, "handling /rules command",
		slog.Int64("chat_id", message.Chat.ID),
		slog.Int("thread_id", message.MessageThreadID),
	)

	_, err := ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:          message.Chat.ChatID(),
		MessageThreadID: message.MessageThreadID,
		ParseMode:       telego.ModeHTML,
		Text:            messages.Rules,
	})
	if err != nil {
		log.Error("Sending rules message failed", slog.Any("error", err))
	}

	return nil
}

func (h *handler) usefulInfo(ctx *th.Context, message telego.Message) error {
	log := logger.FromContext(ctx)

	log = log.With(slog.Group("user",
		slog.String("username", message.From.Username),
		slog.Int64("id", message.From.ID),
	))

	log.Log(ctx, logger.LevelTrace, "handling /useful command",
		slog.Int64("chat_id", message.Chat.ID),
		slog.Int("thread_id", message.MessageThreadID),
	)

	_, err := ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:          message.Chat.ChatID(),
		MessageThreadID: message.MessageThreadID,
		ParseMode:       telego.ModeHTML,
		Text:            messages.Resources,
	})
	if err != nil {
		log.Error("Sending useful info message failed", slog.Any("error", err))
	}

	return nil
}

func (h *handler) help(ctx *th.Context, message telego.Message) error {
	log := logger.FromContext(ctx)

	log = log.With(slog.Group("user",
		slog.String("username", message.From.Username),
		slog.Int64("id", message.From.ID),
	))

	log.Log(ctx, logger.LevelTrace, "handling /help command",
		slog.Int64("chat_id", message.Chat.ID),
		slog.Int("thread_id", message.MessageThreadID),
	)

	_, err := ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:          message.Chat.ChatID(),
		MessageThreadID: message.MessageThreadID,
		ParseMode:       telego.ModeHTML,
		Text:            fmt.Sprintf(messages.Help, message.From.FirstName, viper.GetString("admin-username")),
	})
	if err != nil {
		log.Error("Sending help message failed", slog.Any("error", err))
	}

	return nil
}
