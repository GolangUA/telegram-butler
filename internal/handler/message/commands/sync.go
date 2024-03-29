package commands

import (
	"slices"

	"github.com/mymmrac/telego"
)

func Sync(b *telego.Bot) error {
	commands, err := b.GetMyCommands(&telego.GetMyCommandsParams{})
	if err != nil {
		return err
	}

	if !slices.Equal(commands, allCommands) {
		return b.SetMyCommands(&telego.SetMyCommandsParams{
			Commands: allCommands,
		})
	}

	return nil
}
