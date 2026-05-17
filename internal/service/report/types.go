package report

import (
	"context"
	"time"

	"github.com/GolangUA/telegram-butler/internal/entity"
)

const (
	Quorum       = 5
	VoteWindow   = 3 * time.Minute
	MuteDuration = 3 * time.Hour
)

// VoteAction is sent from the callback handler to the vote goroutine.
type VoteAction struct {
	Voter entity.Voter
}

// VoteResult is returned from the goroutine to the callback handler.
type VoteResult struct {
	Vote     *entity.Vote
	Finished bool
	Error    error
}

// ExpireFunc is invoked from the goroutine on timer expiry.
// This is the only path where the goroutine itself talks to Telegram —
// no callback handler exists at that moment to do it on its behalf.
type ExpireFunc func(ctx context.Context, vote *entity.Vote)

// Repository is the persistence layer the service depends on.
type Repository interface {
	Create(ctx context.Context, vote *entity.Vote) error
	Active(ctx context.Context, key entity.VoteKey) (*entity.Vote, error)
	AddVoter(ctx context.Context, key entity.VoteKey, voter entity.Voter) (*entity.Vote, error)
	SetStatus(ctx context.Context, key entity.VoteKey, status entity.VoteStatus) error
	ListActive(ctx context.Context) ([]*entity.Vote, error)
}
