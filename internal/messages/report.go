package messages

const (
	ReportVoteHeader   = "⚠️ Голосування за mute для %s (3 хв)"
	ReportVoteSection  = "🔇 Голосували (%d/%d):\n%s"
	ReportButtonMute   = "Mute (%d/%d)"
	ReportVoteMuted    = "🔇 %s отримав mute на %s (%d/%d)"
	ReportVoteAdminsCC = "Cc: %s"
	ReportVoteExpired  = "⏰ Голосування завершено — недостатньо голосів"

	ReportToastAlreadyVoted = "Ви вже проголосували"
	ReportToastTargetVoting = "Ви не можете голосувати"

	ReportInfraError = "⚠️ Сервіс тимчасово недоступний, спробуйте пізніше"
)
