package report

import (
	"context"
	"slices"
	"sync"

	"github.com/GolangUA/telegram-butler/internal/entity"
)

// fakeRepo is an in-memory Repository used by coordinator and service tests
// that need hermetic state without a Firestore emulator.
type fakeRepo struct {
	mu    sync.Mutex
	votes map[entity.VoteKey]*entity.Vote
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{votes: make(map[entity.VoteKey]*entity.Vote)}
}

func (r *fakeRepo) Create(_ context.Context, vote *entity.Vote) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.votes[vote.Key()] = vote

	return nil
}

func (r *fakeRepo) GetActive(_ context.Context, key entity.VoteKey) (*entity.Vote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	vote, ok := r.votes[key]
	if !ok || vote.Status != entity.VoteStatusActive {
		return nil, entity.ErrVoteNotFound
	}

	return vote, nil
}

func (r *fakeRepo) AddVoter(_ context.Context, key entity.VoteKey, voter entity.Voter) (*entity.Vote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	vote, ok := r.votes[key]
	if !ok || vote.Status != entity.VoteStatusActive {
		return nil, entity.ErrVoteNotFound
	}

	if !slices.ContainsFunc(vote.Voters, func(v entity.Voter) bool { return v.ID == voter.ID }) {
		vote.Voters = append(vote.Voters, voter)
	}

	return vote, nil
}

func (r *fakeRepo) SetStatus(_ context.Context, key entity.VoteKey, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	vote, ok := r.votes[key]
	if !ok {
		return entity.ErrVoteNotFound
	}

	vote.Status = status

	return nil
}

func (r *fakeRepo) ListActive(_ context.Context) ([]*entity.Vote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var active []*entity.Vote

	for _, v := range r.votes {
		if v.Status == entity.VoteStatusActive {
			active = append(active, v)
		}
	}

	return active, nil
}
