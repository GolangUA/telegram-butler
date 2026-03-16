package messages

const (
	MuteNotification = "🔇 %s muted by %s for %s"

	MuteError = "⚠️ Command error\n\n" +
		"Command: <code>%s</code>\n" +
		"Error: %s\n\n" +
		"Usage: <code>/m &lt;duration&gt; [reason]</code> (reply to a message)\n" +
		"Duration: m/min/minute, h, d, w, mo/month (e.g. 30m, 1h, 2d, 1w, 1mo)\n\n" +
		"<i>This message will be deleted after 15 seconds</i>"
)
