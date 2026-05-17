package mute

import (
	"fmt"
	"html"

	"github.com/mymmrac/telego"

	"github.com/GolangUA/telegram-butler/internal/duration"
	"github.com/GolangUA/telegram-butler/internal/mention"
	"github.com/GolangUA/telegram-butler/internal/messages"
)

// formatMuteNotification builds the HTML-formatted chat notification posted
// after a successful /mute.
func formatMuteNotification(target, caller *telego.User, cmd *command) string {
	notification := fmt.Sprintf(messages.MuteNotification,
		mention.User(target), mention.User(caller), duration.Format(cmd.Duration))

	if cmd.Reason != "" {
		notification += fmt.Sprintf("\n%s: %s", messages.MuteReason, html.EscapeString(cmd.Reason))
	}

	return notification
}

// formatMuteError builds the HTML-formatted error reply shown when /mute is
// rejected. The caller is responsible for the auto-delete lifecycle.
func formatMuteError(cmdText, errText string) string {
	return fmt.Sprintf(messages.MuteError, html.EscapeString(cmdText), html.EscapeString(errText))
}
