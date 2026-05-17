package report

import (
	"context"
	"fmt"
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
func (s *Service) CastVote(ctx context.Context, key entity.VoteKey, voter entity.Voter) (*VoteResult, error) {
	return s.coordinator.Vote(ctx, key, voter)
}

// ActiveVote returns the active vote for a target, or entity.ErrVoteNotFound.
func (s *Service) ActiveVote(ctx context.Context, key entity.VoteKey) (*entity.Vote, error) {
	return s.repo.Active(ctx, key)
}

// Reconcile loads active votes from persistence and re-spawns a coordinator
// goroutine for each. Intended to run once at startup. Past-deadline votes
// are handled naturally — the goroutine's context.WithDeadline fires
// immediately and the existing expiry path runs.
func (s *Service) Reconcile(ctx context.Context) error {
	votes, err := s.repo.ListActive(ctx)
	if err != nil {
		return fmt.Errorf("list active: %w", err)
	}

	log := slog.Default()
	log.Info("Reconciling active votes", slog.Int("count", len(votes)))

	for _, v := range votes {
		s.coordinator.Start(v)

		log.Info("Vote resumed",
			slog.Int64("chat_id", v.ChatID),
			slog.Int64("target_id", v.TargetUserID),
			slog.Time("expires_at", v.ExpiresAt),
		)
	}

	return nil
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
