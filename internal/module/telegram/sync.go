package telegram

import (
	"github.com/mymmrac/telego"
	"github.com/spf13/viper"
)

func syncName(b *telego.Bot) error {
	my, err := b.GetMyName(&telego.GetMyNameParams{})
	if err != nil {
		return err
	}

	if actualName := viper.GetString("bot-name"); my.Name != actualName {
		return b.SetMyName(&telego.SetMyNameParams{
			Name: actualName,
		})
	}

	return nil
}
