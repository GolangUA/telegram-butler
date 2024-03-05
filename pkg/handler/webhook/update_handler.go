package webhook

import (
	"context"
	"errors"
	"log/slog"

	"github.com/mymmrac/telego"
)

func handleUpdate(ctx context.Context, update telego.Update) error {
	// TODO: Remove
	if update.UpdateID == 0 {
		return errors.New("invalid update")
	}

	slog.InfoContext(ctx, "update", "id", update.UpdateID)
	return nil
}
