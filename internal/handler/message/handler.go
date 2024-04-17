package message

import (
	"context"
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
	bh.HandleMessageCtx(h.rules, th.CommandEqual(commands.SendRules))
	bh.HandleMessageCtx(h.usefulInfo, th.CommandEqual(commands.SendUsefulInfo))
	bh.HandleMessageCtx(h.help, th.CommandEqual(commands.SendHelp))
}

type handler struct{}

func (h *handler) rules(ctx context.Context, bot *telego.Bot, message telego.Message) {
	log := logger.FromContext(ctx)

	log = log.With(slog.Group("user",
		slog.String("username", message.From.Username),
		slog.Int64("id", message.From.ID),
	))

	_, err := bot.SendMessage(&telego.SendMessageParams{
		ChatID:    message.Chat.ChatID(),
		ParseMode: telego.ModeHTML,
		Text:      messages.Rules,
	})
	if err != nil {
		log.Error("Sending rules message failed", slog.Any("error", err))
	}
}

func (h *handler) usefulInfo(ctx context.Context, bot *telego.Bot, message telego.Message) {
	log := logger.FromContext(ctx)

	log = log.With(slog.Group("user",
		slog.String("username", message.From.Username),
		slog.Int64("id", message.From.ID),
	))

	_, err := bot.SendMessage(&telego.SendMessageParams{
		ChatID:    message.Chat.ChatID(),
		ParseMode: telego.ModeHTML,
		Text:      messages.Resources,
	})
	if err != nil {
		log.Error("Sending useful info message failed", slog.Any("error", err))
	}
}

func (h *handler) help(ctx context.Context, bot *telego.Bot, message telego.Message) {
	log := logger.FromContext(ctx)

	log = log.With(slog.Group("user",
		slog.String("username", message.From.Username),
		slog.Int64("id", message.From.ID),
	))

	_, err := bot.SendMessage(&telego.SendMessageParams{
		ChatID:    message.Chat.ChatID(),
		ParseMode: telego.ModeHTML,
		Text:      fmt.Sprintf(messages.Help, message.From.FirstName, viper.GetString("admin-username")),
	})
	if err != nil {
		log.Error("Sending help message failed", slog.Any("error", err))
	}
}
