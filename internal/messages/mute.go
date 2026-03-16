package messages

const (
	MuteNotification = "🔇 @%s muted by @%s for %s"

	MuteError = "⚠️ Command error\n\n" +
		"Command: `%s`\n" +
		"Error: %s\n\n" +
		"Usage: `/m <duration> [reason]` \\(reply to a message\\)\n" +
		// https://core.telegram.org/bots/api#markdownv2-style
		// In all other places characters
		// '_', '*', '[', ']', '(', ')', '~', '`', '>', '#',
		// '+', '-', '=', '|', '{', '}', '.', '!'
		// must be escaped with the preceding character '\'.
		"Duration: m, h, d, w, mo \\(e\\.g\\. 30m, 1h, 2d, 1w, 1mo\\)\n\n" +
		"_This message will be deleted after 15 seconds_"
)
