package commands

import (
	"context"
	"slices"

	"github.com/mymmrac/telego"
)

func Sync(ctx context.Context, b *telego.Bot) error {
	err := syncPublicCommands(ctx, b)
	if err != nil {
		return err
	}

	return syncAdminCommands(ctx, b)
}

func syncPublicCommands(ctx context.Context, b *telego.Bot) error {
	commands, err := b.GetMyCommands(ctx, &telego.GetMyCommandsParams{})
	if err != nil {
		return err
	}

	if !slices.Equal(commands, publicCommands) {
		return b.SetMyCommands(ctx, &telego.SetMyCommandsParams{
			Commands: publicCommands,
		})
	}

	return nil
}

func syncAdminCommands(ctx context.Context, b *telego.Bot) error {
	scope := &telego.BotCommandScopeAllChatAdministrators{Type: "all_chat_administrators"}

	commands, err := b.GetMyCommands(ctx, &telego.GetMyCommandsParams{
		Scope: scope,
	})
	if err != nil {
		return err
	}

	if !slices.Equal(commands, adminCommands) {
		return b.SetMyCommands(ctx, &telego.SetMyCommandsParams{
			Commands: adminCommands,
			Scope:    scope,
		})
	}

	return nil
}
