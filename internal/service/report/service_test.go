package report

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/GolangUA/telegram-butler/internal/entity"
	"github.com/GolangUA/telegram-butler/internal/repository/memory"
)

func TestService_ActiveVote_NotFound(t *testing.T) {
	t.Parallel()

	s := NewService(memory.NewVoteRepository(), nil)

	_, err := s.ActiveVote(context.Background(), entity.VoteKey{ChatID: 1, TargetUserID: 2})
	if !errors.Is(err, entity.ErrVoteNotFound) {
		t.Errorf("err = %v, want ErrVoteNotFound", err)
	}
}

func TestService_Reconcile_RespawnsActiveOnly(t *testing.T) {
	t.Parallel()

	repo := memory.NewVoteRepository()
	ctx := context.Background()
	now := time.Now()

	// One vote stays active (target 2), one gets muted (target 3).
	active := &entity.Vote{
		ChatID:       1,
		TargetUserID: 2,
		Voters:       []entity.Voter{{ID: 100}},
		Status:       entity.VoteStatusActive,
		CreatedAt:    now,
		ExpiresAt:    now.Add(time.Minute),
	}
	inactive := &entity.Vote{
		ChatID:       1,
		TargetUserID: 3,
		Voters:       []entity.Voter{{ID: 100}},
		Status:       entity.VoteStatusActive,
		CreatedAt:    now,
		ExpiresAt:    now.Add(time.Minute),
	}

	err := repo.Create(ctx, active)
	if err != nil {
		t.Fatal(err)
	}

	err = repo.Create(ctx, inactive)
	if err != nil {
		t.Fatal(err)
	}

	err = repo.SetStatus(ctx, 1, 3, entity.VoteStatusMuted)
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(repo, nil)

	err = s.Reconcile(ctx)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	// The resumed active vote should accept a new voter.
	_, err = s.CastVote(ctx, entity.VoteKey{ChatID: 1, TargetUserID: 2}, entity.Voter{ID: 200})
	if err != nil {
		t.Errorf("CastVote on resumed vote: %v", err)
	}

	// The muted vote should NOT have a goroutine registered.
	_, err = s.CastVote(ctx, entity.VoteKey{ChatID: 1, TargetUserID: 3}, entity.Voter{ID: 200})
	if !errors.Is(err, entity.ErrVoteNotFound) {
		t.Errorf("CastVote on muted vote = %v, want ErrVoteNotFound", err)
	}
}
