package report

import (
	"context"
	"log/slog"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/GolangUA/telegram-butler/internal/entity"
	"github.com/GolangUA/telegram-butler/internal/messages"
	"github.com/GolangUA/telegram-butler/internal/module/logger"
)

// Service is the business-logic layer for community vote-mute. It owns the
// repository, the per-vote coordinator goroutines, and a bot reference used
// solely for the expiry-side Telegram edit (when no callback handler is alive
// to do it on our behalf).
type Service struct {
	repo        Repository
	bot         *telego.Bot
	coordinator *Coordinator
}

func NewService(repo Repository, bot *telego.Bot) *Service {
	s := &Service{repo: repo, bot: bot}
	s.coordinator = NewCoordinator(repo, s.notifyExpired)

	return s
}

// StartVote persists the vote and starts the per-vote coordinator goroutine.
func (s *Service) StartVote(ctx context.Context, vote *entity.Vote) error {
	err := s.repo.Create(ctx, vote)
	if err != nil {
		return err
	}

	s.coordinator.Start(vote)

	return nil
}

// CastVote registers a voter and returns the updated vote state.
// Returns entity.ErrVoteNotFound if the vote is no longer active.
func (s *Service) CastVote(ctx context.Context, chatID, targetUserID int64, voter entity.Voter) (*VoteResult, error) {
	return s.coordinator.Vote(ctx, chatID, targetUserID, voter)
}

// ActiveVote returns the active vote for a target, or entity.ErrVoteNotFound.
func (s *Service) ActiveVote(ctx context.Context, chatID, targetUserID int64) (*entity.Vote, error) {
	return s.repo.GetActive(ctx, chatID, targetUserID)
}

// notifyExpired runs in goroutine context — no callback handler is alive,
// so the service edits the Telegram message itself.
func (s *Service) notifyExpired(ctx context.Context, vote *entity.Vote) {
	slog.Default().Log(ctx, logger.LevelTrace, "vote expired without quorum",
		slog.Int64("chat_id", vote.ChatID),
		slog.Int64("target_id", vote.TargetUserID),
		slog.Int("count", len(vote.Voters)),
		slog.Int("quorum", Quorum),
	)

	_, err := s.bot.EditMessageText(ctx, &telego.EditMessageTextParams{
		ChatID:    tu.ID(vote.ChatID),
		MessageID: vote.MessageID,
		Text:      messages.ReportVoteExpired,
		ParseMode: telego.ModeHTML,
	})
	if err != nil {
		slog.Error("notifyExpired: edit message failed", slog.Any("error", err))
		return
	}

	slog.Default().Log(ctx, logger.LevelTrace, "expiry message posted",
		slog.Int("message_id", vote.MessageID),
	)
}
