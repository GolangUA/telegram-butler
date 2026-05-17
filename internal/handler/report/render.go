package report

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/GolangUA/telegram-butler/internal/duration"
	"github.com/GolangUA/telegram-butler/internal/entity"
	"github.com/GolangUA/telegram-butler/internal/mention"
	"github.com/GolangUA/telegram-butler/internal/messages"
	reportsvc "github.com/GolangUA/telegram-butler/internal/service/report"
)

// renderVoteMessage builds the active-vote text. If target is nil, it falls
// back to the cached display name on the vote (used during expiry / edits when
// we no longer have the original *telego.User).
func renderVoteMessage(vote *entity.Vote, target *telego.User) string {
	var targetMention string
	if target != nil {
		targetMention = mention.User(target)
	} else {
		targetMention = mention.ByID(vote.TargetUserID, vote.TargetName)
	}

	voterLines := make([]string, 0, len(vote.Voters))
	for _, v := range vote.Voters {
		voterLines = append(voterLines, "- "+mention.HTML(v.ID, v.FirstName, v.Username))
	}

	header := fmt.Sprintf(messages.ReportVoteHeader, targetMention)
	section := fmt.Sprintf(
		messages.ReportVoteSection,
		len(vote.Voters),
		reportsvc.Quorum,
		strings.Join(voterLines, "\n"),
	)

	return header + "\n\n" + section
}

func renderMuteResult(vote *entity.Vote, admins []telego.ChatMember) string {
	target := mention.ByID(vote.TargetUserID, vote.TargetName)
	body := fmt.Sprintf(
		messages.ReportVoteMuted,
		target,
		duration.Format(reportsvc.MuteDuration),
		len(vote.Voters),
		reportsvc.Quorum,
	)

	if mentions := adminMentions(admins); len(mentions) > 0 {
		body += "\n\n" + fmt.Sprintf(messages.ReportVoteAdminsCC, strings.Join(mentions, " "))
	}

	return body
}

func adminMentions(admins []telego.ChatMember) []string {
	mentions := make([]string, 0, len(admins))
	for _, a := range admins {
		u := a.MemberUser()
		if u.IsBot {
			continue
		}

		mentions = append(mentions, mention.User(&u))
	}

	return mentions
}

func buildVoteKeyboard(vote *entity.Vote) *telego.InlineKeyboardMarkup {
	return tu.InlineKeyboard(tu.InlineKeyboardRow(telego.InlineKeyboardButton{
		Text:         fmt.Sprintf(messages.ReportButtonMute, len(vote.Voters), reportsvc.Quorum),
		CallbackData: encodeCallbackData(vote.ChatID, vote.TargetUserID),
	}))
}
