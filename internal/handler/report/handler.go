package report

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/GolangUA/telegram-butler/internal/duration"
	"github.com/GolangUA/telegram-butler/internal/entity"
	"github.com/GolangUA/telegram-butler/internal/handler/message/commands"
	"github.com/GolangUA/telegram-butler/internal/mention"
	"github.com/GolangUA/telegram-butler/internal/messages"
	"github.com/GolangUA/telegram-butler/internal/module/logger"
	"github.com/GolangUA/telegram-butler/internal/module/telegram"
)

const callbackPrefix = "report_vote_"

func Register(bh *th.BotHandler, bot *telego.Bot, repo Repository) {
	h := &handler{bot: bot, repo: repo}
	h.coordinator = NewCoordinator(repo, h.expireVote)

	bh.HandleMessage(h.handleReport, th.CommandEqual(commands.Report))

	// All vote callbacks share the prefix — group them so the predicate is declared once.
	voteGroup := bh.Group(th.CallbackDataPrefix(callbackPrefix))
	voteGroup.HandleCallbackQuery(h.handleVote)
}

type handler struct {
	bot         *telego.Bot
	repo        Repository
	coordinator *Coordinator
}

func (h *handler) handleReport(ctx *th.Context, message telego.Message) error {
	log := logger.FromContext(ctx).With(slog.Group("user",
		slog.String("username", message.From.Username),
		slog.Int64("id", message.From.ID),
	))

	if err := h.validateReport(ctx, log, message); err != nil {
		log.Log(ctx, logger.LevelTrace, "report rejected", slog.Any("error", err))
		h.deleteSilent(ctx, message)
		return nil
	}

	target := message.ReplyToMessage.From
	now := time.Now()

	reporter := voterFromUser(message.From)
	vote := &entity.Vote{
		ChatID:       message.Chat.ID,
		TargetUserID: target.ID,
		TargetName:   mention.DisplayName(target),
		ReporterID:   message.From.ID,
		ThreadID:     message.MessageThreadID,
		Voters:       []entity.Voter{reporter},
		Status:       entity.VoteStatusActive,
		CreatedAt:    now,
		ExpiresAt:    now.Add(VoteWindow),
	}

	if err := h.repo.Create(ctx, vote); err != nil {
		log.Error("Failed to create vote", slog.Any("error", err))
		h.deleteSilent(ctx, message)
		return nil
	}

	sent, err := h.sendVoteMessage(ctx, vote, target)
	if err != nil {
		log.Error("Failed to send vote message", slog.Any("error", err))
		_ = h.repo.SetStatus(ctx, vote.ChatID, vote.TargetUserID, entity.VoteStatusExpired)
		return nil
	}

	vote.MessageID = sent.MessageID

	h.coordinator.Start(vote)

	log.Info("Vote started",
		slog.Int64("target_id", target.ID),
		slog.String("target", target.Username),
	)
	return nil
}

func (h *handler) handleVote(ctx *th.Context, query telego.CallbackQuery) error {
	log := logger.FromContext(ctx).With(slog.Group("user",
		slog.String("username", query.From.Username),
		slog.Int64("id", query.From.ID),
	))

	log.Log(ctx, logger.LevelTrace, "processing vote callback", slog.String("data", query.Data))

	chatID, targetUserID, err := parseCallbackData(query.Data)
	if err != nil {
		log.Error("Failed to parse callback data", slog.Any("error", err), slog.String("data", query.Data))
		return nil
	}

	if query.From.ID == targetUserID {
		log.Log(ctx, logger.LevelTrace, "vote rejected: target voting on self",
			slog.Int64("target_id", targetUserID),
		)
		h.answerToast(ctx, query.ID, messages.ReportToastTargetVoting)
		return nil
	}

	current, err := h.repo.GetActive(ctx, chatID, targetUserID)
	if err != nil {
		log.Error("Failed to get active vote", slog.Any("error", err))
		h.answerToast(ctx, query.ID, "")
		return nil
	}

	alreadyVoted := slices.ContainsFunc(current.Voters, func(v entity.Voter) bool { return v.ID == query.From.ID })
	if alreadyVoted {
		log.Log(ctx, logger.LevelTrace, "vote rejected: already voted",
			slog.Int64("target_id", targetUserID),
		)
		h.answerToast(ctx, query.ID, messages.ReportToastAlreadyVoted)
		return nil
	}

	result, err := h.coordinator.Vote(ctx, chatID, targetUserID, voterFromUser(&query.From))
	if err != nil {
		log.Error("Failed to register vote", slog.Any("error", err))
		h.answerToast(ctx, query.ID, "")
		return nil
	}

	if result.Error != nil {
		log.Error("Vote processing failed", slog.Any("error", result.Error))
		h.answerToast(ctx, query.ID, "")
		return nil
	}

	h.answerToast(ctx, query.ID, "")

	log.Log(ctx, logger.LevelTrace, "vote registered",
		slog.Int64("target_id", targetUserID),
		slog.Int("count", len(result.Vote.Voters)),
		slog.Int("quorum", Quorum),
		slog.Bool("finished", result.Finished),
	)

	if result.Finished {
		if err := h.applyMute(ctx, result.Vote); err != nil {
			log.Error("Failed to apply mute", slog.Any("error", err))
		}
		return nil
	}

	if err := h.editVoteMessage(ctx, result.Vote); err != nil {
		log.Error("Failed to edit vote message", slog.Any("error", err))
	}

	return nil
}

func (h *handler) expireVote(ctx context.Context, vote *entity.Vote) {
	_, err := h.bot.EditMessageText(ctx, &telego.EditMessageTextParams{
		ChatID:    tu.ID(vote.ChatID),
		MessageID: vote.MessageID,
		Text:      messages.ReportVoteExpired,
		ParseMode: telego.ModeHTML,
	})
	if err != nil {
		slog.Error("expireVote: edit message failed", slog.Any("error", err))
	}
}

func (h *handler) validateReport(ctx *th.Context, log *slog.Logger, message telego.Message) error {
	if message.ReplyToMessage == nil || message.ReplyToMessage.From == nil {
		return errors.New("not a reply")
	}

	target := message.ReplyToMessage.From

	if target.ID == message.From.ID {
		return errors.New("self-report")
	}

	if target.IsBot {
		return errors.New("bot target")
	}

	chatID := message.Chat.ChatID()

	isAdmin, err := h.isAdmin(ctx, chatID, target.ID)
	if err != nil {
		log.Error("Failed to check target admin status", slog.Any("error", err))
		return errors.New("admin check failed")
	}
	if isAdmin {
		return errors.New("target is admin")
	}

	_, err = h.repo.GetActive(ctx, message.Chat.ID, target.ID)
	if err == nil {
		return errors.New("active vote exists")
	}
	if !errors.Is(err, entity.ErrVoteNotFound) {
		log.Error("Failed to check active vote", slog.Any("error", err))
		return errors.New("active vote check failed")
	}

	return nil
}

func (h *handler) applyMute(ctx context.Context, vote *entity.Vote) error {
	err := h.bot.RestrictChatMember(ctx, &telego.RestrictChatMemberParams{
		ChatID:      tu.ID(vote.ChatID),
		UserID:      vote.TargetUserID,
		UntilDate:   time.Now().Add(MuteDuration).Unix(),
		Permissions: telegram.DenyAllPermissions(),
	})
	if err != nil {
		return fmt.Errorf("restrict chat member: %w", err)
	}

	admins, err := h.bot.GetChatAdministrators(ctx, &telego.GetChatAdministratorsParams{
		ChatID: tu.ID(vote.ChatID),
	})
	if err != nil {
		slog.Error("applyMute: failed to fetch admins for tagging", slog.Any("error", err))
		admins = nil
	}

	text := renderMuteResult(vote, admins)

	_, err = h.bot.EditMessageText(ctx, &telego.EditMessageTextParams{
		ChatID:    tu.ID(vote.ChatID),
		MessageID: vote.MessageID,
		Text:      text,
		ParseMode: telego.ModeHTML,
	})
	if err != nil {
		return fmt.Errorf("edit final message: %w", err)
	}

	return nil
}

func (h *handler) sendVoteMessage(ctx context.Context, vote *entity.Vote, target *telego.User) (*telego.Message, error) {
	keyboard := tu.InlineKeyboard(tu.InlineKeyboardRow(telego.InlineKeyboardButton{
		Text:         fmt.Sprintf(messages.ReportButtonMute, len(vote.Voters), Quorum),
		CallbackData: encodeCallbackData(vote.ChatID, vote.TargetUserID),
	}))

	return h.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:          tu.ID(vote.ChatID),
		MessageThreadID: vote.ThreadID,
		Text:            renderVoteMessage(vote, target),
		ParseMode:       telego.ModeHTML,
		ReplyMarkup:     keyboard,
	})
}

func (h *handler) editVoteMessage(ctx context.Context, vote *entity.Vote) error {
	keyboard := tu.InlineKeyboard(tu.InlineKeyboardRow(telego.InlineKeyboardButton{
		Text:         fmt.Sprintf(messages.ReportButtonMute, len(vote.Voters), Quorum),
		CallbackData: encodeCallbackData(vote.ChatID, vote.TargetUserID),
	}))

	_, err := h.bot.EditMessageText(ctx, &telego.EditMessageTextParams{
		ChatID:      tu.ID(vote.ChatID),
		MessageID:   vote.MessageID,
		Text:        renderVoteMessage(vote, nil),
		ParseMode:   telego.ModeHTML,
		ReplyMarkup: keyboard,
	})
	return err
}

func renderVoteMessage(vote *entity.Vote, target *telego.User) string {
	var targetMention string
	if target != nil {
		targetMention = mention.User(target)
	} else {
		targetMention = mention.ByID(vote.TargetUserID, vote.TargetName)
	}

	var voterLines []string
	for _, v := range vote.Voters {
		voterLines = append(voterLines, "- "+mention.HTML(v.ID, v.FirstName, v.Username))
	}

	return fmt.Sprintf(messages.ReportVoteHeader, targetMention) +
		"\n\n" +
		fmt.Sprintf(messages.ReportVoteSection, len(vote.Voters), Quorum, strings.Join(voterLines, "\n"))
}

func renderMuteResult(vote *entity.Vote, admins []telego.ChatMember) string {
	target := mention.ByID(vote.TargetUserID, vote.TargetName)
	body := fmt.Sprintf(messages.ReportVoteMuted, target, duration.Format(MuteDuration), len(vote.Voters), Quorum)

	mentions := adminMentions(admins)
	if len(mentions) > 0 {
		body += "\n\n" + fmt.Sprintf(messages.ReportVoteAdminsCC, strings.Join(mentions, " "))
	}
	return body
}

func adminMentions(admins []telego.ChatMember) []string {
	var mentions []string
	for _, a := range admins {
		u := a.MemberUser()
		if u.IsBot {
			continue
		}
		mentions = append(mentions, mention.User(&u))
	}
	return mentions
}

func (h *handler) isAdmin(ctx *th.Context, chatID telego.ChatID, userID int64) (bool, error) {
	member, err := ctx.Bot().GetChatMember(ctx, &telego.GetChatMemberParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return false, fmt.Errorf("get chat member: %w", err)
	}

	status := member.MemberStatus()
	return status == telego.MemberStatusCreator || status == telego.MemberStatusAdministrator, nil
}

func (h *handler) deleteSilent(ctx *th.Context, message telego.Message) {
	_ = ctx.Bot().DeleteMessage(ctx, tu.Delete(message.Chat.ChatID(), message.MessageID))
}

func (h *handler) answerToast(ctx *th.Context, queryID, text string) {
	params := tu.CallbackQuery(queryID)
	if text != "" {
		params = params.WithText(text)
	}
	_ = ctx.Bot().AnswerCallbackQuery(ctx, params)
}

func voterFromUser(u *telego.User) entity.Voter {
	return entity.Voter{
		ID:        u.ID,
		FirstName: u.FirstName,
		Username:  u.Username,
	}
}

func encodeCallbackData(chatID, targetUserID int64) string {
	return fmt.Sprintf("%s%d_%d", callbackPrefix, chatID, targetUserID)
}

func parseCallbackData(data string) (chatID, targetUserID int64, err error) {
	if !strings.HasPrefix(data, callbackPrefix) {
		return 0, 0, fmt.Errorf("missing prefix %q", callbackPrefix)
	}

	rest := strings.TrimPrefix(data, callbackPrefix)
	parts := strings.Split(rest, "_")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected 2 parts, got %d", len(parts))
	}

	chatID, err = parseInt64(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid chatID: %w", err)
	}
	targetUserID, err = parseInt64(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid targetUserID: %w", err)
	}
	return chatID, targetUserID, nil
}

func parseInt64(s string) (int64, error) {
	var n int64
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
