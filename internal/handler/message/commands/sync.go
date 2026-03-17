package commands

import (
	"context"
	"slices"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
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
	commands, err := b.GetMyCommands(ctx, &telego.GetMyCommandsParams{
		Scope: tu.ScopeAllChatAdministrators(),
	})
	if err != nil {
		return err
	}

	if !slices.Equal(commands, adminCommands) {
		return b.SetMyCommands(ctx, &telego.SetMyCommandsParams{
			Commands: adminCommands,
			Scope:    tu.ScopeAllChatAdministrators(),
		})
	}

	return nil
}
