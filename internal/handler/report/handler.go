package report

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/GolangUA/telegram-butler/internal/entity"
	"github.com/GolangUA/telegram-butler/internal/handler/message/commands"
	"github.com/GolangUA/telegram-butler/internal/mention"
	"github.com/GolangUA/telegram-butler/internal/messages"
	"github.com/GolangUA/telegram-butler/internal/module/logger"
	"github.com/GolangUA/telegram-butler/internal/module/telegram"
	reportsvc "github.com/GolangUA/telegram-butler/internal/service/report"
)

const callbackPrefix = "report_vote_"

// errReportInfra distinguishes infrastructure failures (Firestore unreachable,
// Telegram admin API errors) from caller-side validation errors. Validation
// failures stay silent (delete the caller's message); infra failures get a
// visible chat reply so the user knows the bot is degraded.
var errReportInfra = errors.New("infra failure")

// Service is the abstract behavior the handler depends on. The concrete
// implementation lives in internal/service/report; the handler only needs
// these three methods, so it accepts an interface instead of the struct.
type Service interface {
	StartVote(ctx context.Context, vote *entity.Vote) error
	ActiveVote(ctx context.Context, chatID, targetUserID int64) (*entity.Vote, error)
	CastVote(ctx context.Context, chatID, targetUserID int64, voter entity.Voter) (*reportsvc.VoteResult, error)
}

func Register(bh *th.BotHandler, bot *telego.Bot, svc Service) {
	h := &handler{bot: bot, svc: svc}

	bh.HandleMessage(h.handleReport, th.CommandEqual(commands.Report))

	// All vote callbacks share the prefix — group them so the predicate is declared once.
	voteGroup := bh.Group(th.CallbackDataPrefix(callbackPrefix))
	voteGroup.HandleCallbackQuery(h.handleVote)
}

type handler struct {
	bot *telego.Bot
	svc Service
}

func (h *handler) handleReport(ctx *th.Context, message telego.Message) error {
	log := logger.FromContext(ctx).With(slog.Group("user",
		slog.String("username", message.From.Username),
		slog.Int64("id", message.From.ID),
	))

	err := h.validateReport(ctx, log, message)
	if err != nil {
		if errors.Is(err, errReportInfra) {
			log.Error("Report infra failure", slog.Any("error", err))
			h.replyInfraError(ctx, message)
		} else {
			log.Log(ctx, logger.LevelTrace, "report rejected", slog.Any("error", err))
			h.deleteSilent(ctx, message)
		}

		return nil
	}

	target := message.ReplyToMessage.From
	now := time.Now()

	vote := &entity.Vote{
		ChatID:       message.Chat.ID,
		TargetUserID: target.ID,
		TargetName:   mention.DisplayName(target),
		ReporterID:   message.From.ID,
		ThreadID:     message.MessageThreadID,
		Voters:       []entity.Voter{voterFromUser(message.From)},
		Status:       entity.VoteStatusActive,
		CreatedAt:    now,
		ExpiresAt:    now.Add(reportsvc.VoteWindow),
	}

	sent, err := h.sendVoteMessage(ctx, vote, target)
	if err != nil {
		log.Error("Failed to send vote message", slog.Any("error", err))
		return nil
	}

	log.Log(ctx, logger.LevelTrace, "vote message posted",
		slog.Int("message_id", sent.MessageID),
		slog.Int64("target_id", target.ID),
	)

	vote.MessageID = sent.MessageID

	err = h.svc.StartVote(ctx, vote)
	if err != nil {
		log.Error("Failed to start vote", slog.Any("error", err))
		// Best-effort cleanup — remove the vote message we just posted.
		_ = h.bot.DeleteMessage(ctx, tu.Delete(message.Chat.ChatID(), sent.MessageID))

		return nil
	}

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
		log.Error("Failed to parse callback data",
			slog.Any("error", err),
			slog.String("data", query.Data),
		)

		return nil
	}

	if query.From.ID == targetUserID {
		log.Log(ctx, logger.LevelTrace, "vote rejected: target voting on self",
			slog.Int64("target_id", targetUserID),
		)
		h.answerToast(ctx, query.ID, messages.ReportToastTargetVoting)

		return nil
	}

	current, err := h.svc.ActiveVote(ctx, chatID, targetUserID)
	if err != nil {
		log.Error("Failed to get active vote", slog.Any("error", err))
		h.answerToast(ctx, query.ID, "")

		return nil
	}

	if slices.ContainsFunc(current.Voters, func(v entity.Voter) bool { return v.ID == query.From.ID }) {
		log.Log(ctx, logger.LevelTrace, "vote rejected: already voted",
			slog.Int64("target_id", targetUserID),
		)
		h.answerToast(ctx, query.ID, messages.ReportToastAlreadyVoted)

		return nil
	}

	result, err := h.svc.CastVote(ctx, chatID, targetUserID, voterFromUser(&query.From))
	if err != nil {
		log.Error("Failed to register vote", slog.Any("error", err))
		h.answerToast(ctx, query.ID, "")

		return nil
	}

	if result.Error != nil {
		log.Error("Vote processing failed", slog.Any("error", result.Error))
		h.answerToast(ctx, query.ID, "")

		// result.Error is a domain field on VoteResult, not the function's err
		// (which was handled above). Returning nil signals "handled" to telego.
		return nil //nolint:nilerr
	}

	h.answerToast(ctx, query.ID, "")

	log.Log(ctx, logger.LevelTrace, "vote registered",
		slog.Int64("target_id", targetUserID),
		slog.Int("count", len(result.Vote.Voters)),
		slog.Int("quorum", reportsvc.Quorum),
		slog.Bool("finished", result.Finished),
	)

	if result.Finished {
		err = h.applyMute(ctx, result.Vote)
		if err != nil {
			log.Error("Failed to apply mute", slog.Any("error", err))
		}

		return nil
	}

	err = h.editVoteMessage(ctx, result.Vote)
	if err != nil {
		log.Error("Failed to edit vote message", slog.Any("error", err))
		return nil
	}

	log.Log(ctx, logger.LevelTrace, "vote message updated",
		slog.Int("message_id", result.Vote.MessageID),
		slog.Int("count", len(result.Vote.Voters)),
	)

	return nil
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
		return fmt.Errorf("%w: admin check: %w", errReportInfra, err)
	}

	if isAdmin {
		return errors.New("target is admin")
	}

	_, err = h.svc.ActiveVote(ctx, message.Chat.ID, target.ID)
	if err == nil {
		return errors.New("active vote exists")
	}

	if !errors.Is(err, entity.ErrVoteNotFound) {
		log.Error("Failed to check active vote", slog.Any("error", err))
		return fmt.Errorf("%w: active vote check: %w", errReportInfra, err)
	}

	return nil
}

func (h *handler) applyMute(ctx context.Context, vote *entity.Vote) error {
	log := logger.FromContext(ctx)

	err := h.bot.RestrictChatMember(ctx, &telego.RestrictChatMemberParams{
		ChatID:      tu.ID(vote.ChatID),
		UserID:      vote.TargetUserID,
		UntilDate:   time.Now().Add(reportsvc.MuteDuration).Unix(),
		Permissions: telegram.DenyAllPermissions(),
	})
	if err != nil {
		return err
	}

	log.Log(ctx, logger.LevelTrace, "mute applied",
		slog.Int64("target_id", vote.TargetUserID),
		slog.String("duration", reportsvc.MuteDuration.String()),
	)

	admins, err := h.bot.GetChatAdministrators(ctx, &telego.GetChatAdministratorsParams{
		ChatID: tu.ID(vote.ChatID),
	})
	if err != nil {
		log.Error("applyMute: failed to fetch admins for tagging", slog.Any("error", err))

		admins = nil
	}

	_, err = h.bot.EditMessageText(ctx, &telego.EditMessageTextParams{
		ChatID:    tu.ID(vote.ChatID),
		MessageID: vote.MessageID,
		Text:      renderMuteResult(vote, admins),
		ParseMode: telego.ModeHTML,
	})
	if err != nil {
		return err
	}

	log.Log(ctx, logger.LevelTrace, "mute result message posted",
		slog.Int("message_id", vote.MessageID),
		slog.Int("admin_mentions", len(admins)),
	)

	return nil
}

func (h *handler) sendVoteMessage(
	ctx context.Context, vote *entity.Vote, target *telego.User,
) (*telego.Message, error) {
	return h.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:          tu.ID(vote.ChatID),
		MessageThreadID: vote.ThreadID,
		Text:            renderVoteMessage(vote, target),
		ParseMode:       telego.ModeHTML,
		ReplyMarkup:     buildVoteKeyboard(vote),
	})
}

func (h *handler) editVoteMessage(ctx context.Context, vote *entity.Vote) error {
	_, err := h.bot.EditMessageText(ctx, &telego.EditMessageTextParams{
		ChatID:      tu.ID(vote.ChatID),
		MessageID:   vote.MessageID,
		Text:        renderVoteMessage(vote, nil),
		ParseMode:   telego.ModeHTML,
		ReplyMarkup: buildVoteKeyboard(vote),
	})

	return err
}

func (h *handler) isAdmin(ctx *th.Context, chatID telego.ChatID, userID int64) (bool, error) {
	member, err := ctx.Bot().GetChatMember(ctx, &telego.GetChatMemberParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return false, err
	}

	status := member.MemberStatus()

	return status == telego.MemberStatusCreator || status == telego.MemberStatusAdministrator, nil
}

func (h *handler) replyInfraError(ctx *th.Context, message telego.Message) {
	_, err := ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:          message.Chat.ChatID(),
		MessageThreadID: message.MessageThreadID,
		Text:            messages.ReportInfraError,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: message.MessageID,
		},
	})
	if err != nil {
		logger.FromContext(ctx).Error("Failed to send infra error reply", slog.Any("error", err))
	}
}

func (h *handler) deleteSilent(ctx *th.Context, message telego.Message) {
	err := ctx.Bot().DeleteMessage(ctx, tu.Delete(message.Chat.ChatID(), message.MessageID))
	if err != nil {
		logger.FromContext(ctx).Log(ctx, logger.LevelTrace, "failed to delete report message",
			slog.Int("message_id", message.MessageID),
			slog.Any("error", err),
		)

		return
	}

	logger.FromContext(ctx).Log(ctx, logger.LevelTrace, "report message deleted",
		slog.Int("message_id", message.MessageID),
	)
}

func (h *handler) answerToast(ctx *th.Context, queryID, text string) {
	params := tu.CallbackQuery(queryID)
	if text != "" {
		params = params.WithText(text)
	}

	_ = ctx.Bot().AnswerCallbackQuery(ctx, params)
}
