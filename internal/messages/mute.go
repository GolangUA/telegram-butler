package messages

const (
	MuteNotification = "🔇 %s отримав мут від %s на %s"

	MuteReason = "Причина"

	MuteError = "⚠️ Помилка команди\n\n" +
		"Команда: <code>%s</code>\n" +
		"Помилка: %s\n\n" +
		"Використання: <code>/m &lt;тривалість&gt; [причина]</code> (у відповідь на повідомлення)\n" +
		"Тривалість: m/min/minute, h, d, w, mo/month (напр. 30m, 1h, 2d, 1w, 1mo)\n\n" +
		"<i>Це повідомлення буде видалено через 15 секунд</i>"
)
