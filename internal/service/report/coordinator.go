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

// Coordinator owns the in-flight vote goroutines and their channels.
// Each active vote gets one goroutine that holds vote state in memory and
// owns all repo writes for that vote, so callback handlers can vote
// concurrently without races.
type Coordinator struct {
	repo     Repository
	onExpire ExpireFunc
	registry sync.Map // key: entity.VoteKey, value: chan<- VoteAction
}

func NewCoordinator(repo Repository, onExpire ExpireFunc) *Coordinator {
	return &Coordinator{
		repo:     repo,
		onExpire: onExpire,
	}
}

// Start spawns a goroutine that owns the vote until quorum or expiry.
func (c *Coordinator) Start(vote *entity.Vote) {
	in := make(chan VoteAction)
	c.registry.Store(vote.Key(), in)

	slog.Default().Log(context.Background(), logger.LevelTrace, "vote goroutine started",
		slog.Int64("chat_id", vote.ChatID),
		slog.Int64("target_id", vote.TargetUserID),
		slog.Time("expires_at", vote.ExpiresAt),
	)

	go c.run(vote, in)
}

// Vote sends a voter to the matching goroutine and waits for the result.
// Returns entity.ErrVoteNotFound if no goroutine is registered for the vote.
func (c *Coordinator) Vote(ctx context.Context, key entity.VoteKey, voter entity.Voter) (*VoteResult, error) {
	val, ok := c.registry.Load(key)
	if !ok {
		return nil, entity.ErrVoteNotFound
	}

	in, ok := val.(chan VoteAction)
	if !ok {
		return nil, fmt.Errorf("registry: unexpected value type %T for vote %s", val, key)
	}

	// Buffered size 1 so the coordinator's send never blocks even if our
	// context is canceled before we read the result.
	reply := make(chan VoteResult, 1)

	select {
	case in <- VoteAction{Voter: voter, Reply: reply}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case result := <-reply:
		return &result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Coordinator) run(vote *entity.Vote, in <-chan VoteAction) {
	key := vote.Key()
	defer c.registry.Delete(key)

	ctx, cancel := context.WithDeadline(context.Background(), vote.ExpiresAt)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			c.handleExpire(vote)
			return

		case action := <-in:
			updated, err := c.repo.AddVoter(ctx, key, action.Voter)
			if err != nil {
				action.Reply <- VoteResult{Error: err}
				continue
			}

			if len(updated.Voters) >= Quorum {
				// Status stays VoteStatusActive here — the handler commits muted
				// only after Telegram's RestrictChatMember succeeds, so Firestore
				// never claims muted when the user wasn't actually restricted.
				slog.Default().Log(ctx, logger.LevelTrace, "vote reached quorum",
					slog.Int64("chat_id", vote.ChatID),
					slog.Int64("target_id", vote.TargetUserID),
					slog.Int("count", len(updated.Voters)),
				)

				action.Reply <- VoteResult{Vote: updated, Finished: true}

				return
			}

			action.Reply <- VoteResult{Vote: updated, Finished: false}
		}
	}
}

func (c *Coordinator) handleExpire(vote *entity.Vote) {
	ctx, cancel := context.WithTimeout(context.Background(), expireCleanupTimeout)
	defer cancel()

	_ = c.repo.SetStatus(ctx, vote.Key(), entity.VoteStatusExpired)
	c.onExpire(ctx, vote)
}
