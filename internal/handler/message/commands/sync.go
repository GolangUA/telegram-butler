package commands

import (
	"context"
	"slices"

	"github.com/mymmrac/telego"
)

func Sync(ctx context.Context, b *telego.Bot) error {
	commands, err := b.GetMyCommands(ctx, &telego.GetMyCommandsParams{})
	if err != nil {
		return err
	}

	if !slices.Equal(commands, allCommands) {
		return b.SetMyCommands(ctx, &telego.SetMyCommandsParams{
			Commands: allCommands,
		})
	}

	return nil
}
