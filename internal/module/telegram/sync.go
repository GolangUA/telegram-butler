package telegram

import (
	"fmt"

	"github.com/mymmrac/telego"
	"github.com/spf13/viper"
)

func syncInfo(b *telego.Bot) error {
	if err := syncName(b); err != nil {
		return err
	}

	return syncDescription(b)
}

func syncName(b *telego.Bot) error {
	my, err := b.GetMyName(&telego.GetMyNameParams{})
	if err != nil {
		return err
	}

	actualName := viper.GetString("bot-name")
	if my.Name == actualName {
		return nil
	}

	err = b.SetMyName(&telego.SetMyNameParams{
		Name: actualName,
	})
	if err != nil {
		return fmt.Errorf("sync name failed: %w", err)
	}

	return nil
}

func syncDescription(b *telego.Bot) error {
	my, err := b.GetMyDescription(&telego.GetMyDescriptionParams{})
	if err != nil {
		return err
	}

	actualDescription := viper.GetString("bot-description")
	if my.Description == actualDescription {
		return nil
	}

	err = b.SetMyDescription(&telego.SetMyDescriptionParams{
		Description: actualDescription,
	})
	if err != nil {
		return fmt.Errorf("sync description failed: %w", err)
	}

	return nil
}
