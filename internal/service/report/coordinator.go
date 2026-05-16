package report

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/GolangUA/telegram-butler/internal/entity"
	"github.com/GolangUA/telegram-butler/internal/module/logger"
)

const expireCleanupTimeout = 5 * time.Second

type voteChannels struct {
	in  chan VoteAction
	out chan VoteResult
}

// Coordinator owns the in-flight vote goroutines and their channels.
// Each active vote gets one goroutine that holds vote state in memory and
// owns all repo writes for that vote, so callback handlers can vote
// concurrently without races.
type Coordinator struct {
	repo     Repository
	onExpire ExpireFunc
	registry sync.Map // key: voteKey, value: *voteChannels
}

func NewCoordinator(repo Repository, onExpire ExpireFunc) *Coordinator {
	return &Coordinator{
		repo:     repo,
		onExpire: onExpire,
	}
}

// Start spawns a goroutine that owns the vote until quorum or expiry.
func (c *Coordinator) Start(vote *entity.Vote) {
	ch := &voteChannels{
		in:  make(chan VoteAction),
		out: make(chan VoteResult),
	}
	c.registry.Store(voteKey(vote.ChatID, vote.TargetUserID), ch)

	slog.Default().Log(context.Background(), logger.LevelTrace, "vote goroutine started",
		slog.Int64("chat_id", vote.ChatID),
		slog.Int64("target_id", vote.TargetUserID),
		slog.Time("expires_at", vote.ExpiresAt),
	)

	go c.run(vote, ch)
}

// Vote sends a voter to the matching goroutine and waits for the result.
// Returns entity.ErrVoteNotFound if no goroutine is registered for the vote.
func (c *Coordinator) Vote(ctx context.Context, chatID, targetUserID int64, voter entity.Voter) (*VoteResult, error) {
	val, ok := c.registry.Load(voteKey(chatID, targetUserID))
	if !ok {
		return nil, entity.ErrVoteNotFound
	}

	ch, ok := val.(*voteChannels)
	if !ok {
		return nil, fmt.Errorf("registry: unexpected value type %T for vote %d_%d", val, chatID, targetUserID)
	}

	select {
	case ch.in <- VoteAction{Voter: voter}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case result := <-ch.out:
		return &result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Coordinator) run(vote *entity.Vote, ch *voteChannels) {
	key := voteKey(vote.ChatID, vote.TargetUserID)
	defer c.registry.Delete(key)

	ctx, cancel := context.WithDeadline(context.Background(), vote.ExpiresAt)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			c.handleExpire(vote)
			return

		case action := <-ch.in:
			updated, err := c.repo.AddVoter(ctx, vote.ChatID, vote.TargetUserID, action.Voter)
			if err != nil {
				ch.out <- VoteResult{Error: err}
				continue
			}

			if len(updated.Voters) >= Quorum {
				err := c.repo.SetStatus(ctx, vote.ChatID, vote.TargetUserID, entity.VoteStatusMuted)
				if err != nil {
					ch.out <- VoteResult{Error: err}
					continue
				}

				updated.Status = entity.VoteStatusMuted
				slog.Default().Log(ctx, logger.LevelTrace, "vote reached quorum",
					slog.Int64("chat_id", vote.ChatID),
					slog.Int64("target_id", vote.TargetUserID),
					slog.Int("count", len(updated.Voters)),
				)

				ch.out <- VoteResult{Vote: updated, Finished: true}

				return
			}

			ch.out <- VoteResult{Vote: updated, Finished: false}
		}
	}
}

func (c *Coordinator) handleExpire(vote *entity.Vote) {
	ctx, cancel := context.WithTimeout(context.Background(), expireCleanupTimeout)
	defer cancel()

	_ = c.repo.SetStatus(ctx, vote.ChatID, vote.TargetUserID, entity.VoteStatusExpired)
	c.onExpire(ctx, vote)
}

func voteKey(chatID, targetUserID int64) string {
	return fmt.Sprintf("%d_%d", chatID, targetUserID)
}
