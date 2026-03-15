package telegram

import (
	"context"
	"fmt"

	"github.com/mymmrac/telego"
	"github.com/spf13/viper"
)

func syncInfo(ctx context.Context, b *telego.Bot) error {
	err := syncName(ctx, b)
	if err != nil {
		return err
	}

	return syncDescription(ctx, b)
}

func syncName(ctx context.Context, b *telego.Bot) error {
	my, err := b.GetMyName(ctx, &telego.GetMyNameParams{})
	if err != nil {
		return err
	}

	actualName := viper.GetString("bot-name")
	if my.Name == actualName {
		return nil
	}

	err = b.SetMyName(ctx, &telego.SetMyNameParams{
		Name: actualName,
	})
	if err != nil {
		return fmt.Errorf("sync name failed: %w", err)
	}

	return nil
}

func syncDescription(ctx context.Context, b *telego.Bot) error {
	my, err := b.GetMyDescription(ctx, &telego.GetMyDescriptionParams{})
	if err != nil {
		return err
	}

	actualDescription := viper.GetString("bot-description")
	if my.Description == actualDescription {
		return nil
	}

	err = b.SetMyDescription(ctx, &telego.SetMyDescriptionParams{
		Description: actualDescription,
	})
	if err != nil {
		return fmt.Errorf("sync description failed: %w", err)
	}

	return nil
}
