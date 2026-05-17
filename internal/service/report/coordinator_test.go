package report

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/GolangUA/telegram-butler/internal/entity"
)

func TestCoordinator_ReachesQuorum(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()
	onExpire := func(_ context.Context, _ *entity.Vote) {
		t.Error("onExpire should not be called when quorum is reached")
	}

	c := NewCoordinator(repo, onExpire)
	ctx := context.Background()
	now := time.Now()

	// Seed exactly Quorum-1 voters so a single Vote() call lands the quorum.
	// Built dynamically so the test stays correct if Quorum is tuned.
	seedVoters := make([]entity.Voter, 0, Quorum-1)
	for i := range Quorum - 1 {
		seedVoters = append(seedVoters, entity.Voter{
			ID:        int64(100 + i),
			FirstName: fmt.Sprintf("Seed%d", i),
		})
	}

	vote := &entity.Vote{
		ChatID:       1,
		TargetUserID: 2,
		ReporterID:   100,
		Voters:       seedVoters,
		Status:       entity.VoteStatusActive,
		CreatedAt:    now,
		ExpiresAt:    now.Add(time.Minute),
	}

	err := repo.Create(ctx, vote)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	c.Start(vote)

	result, err := c.Vote(ctx, entity.VoteKey{ChatID: 1, TargetUserID: 2}, entity.Voter{ID: 104, FirstName: "D"})
	if err != nil {
		t.Fatalf("Vote: %v", err)
	}

	if !result.Finished {
		t.Error("result.Finished = false, want true")
	}

	if result.Vote.Status != entity.VoteStatusMuted {
		t.Errorf("status = %q, want %q", result.Vote.Status, entity.VoteStatusMuted)
	}

	if len(result.Vote.Voters) != Quorum {
		t.Errorf("voters len = %d, want %d", len(result.Vote.Voters), Quorum)
	}
}

func TestCoordinator_ExpiresWithoutQuorum(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()

	var (
		expireCalled atomic.Bool
		done         = make(chan struct{})
	)

	onExpire := func(_ context.Context, _ *entity.Vote) {
		expireCalled.Store(true)
		close(done)
	}

	c := NewCoordinator(repo, onExpire)
	ctx := context.Background()
	now := time.Now()

	vote := &entity.Vote{
		ChatID:       1,
		TargetUserID: 2,
		ReporterID:   100,
		Voters:       []entity.Voter{{ID: 100, FirstName: "R"}},
		Status:       entity.VoteStatusActive,
		CreatedAt:    now,
		ExpiresAt:    now.Add(50 * time.Millisecond),
	}

	err := repo.Create(ctx, vote)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	c.Start(vote)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("onExpire was not called within 2s")
	}

	if !expireCalled.Load() {
		t.Fatal("onExpire flag not set")
	}

	_, err = repo.GetActive(ctx, entity.VoteKey{ChatID: 1, TargetUserID: 2})
	if !errors.Is(err, entity.ErrVoteNotFound) {
		t.Errorf("GetActive after expire = %v, want ErrVoteNotFound", err)
	}
}
