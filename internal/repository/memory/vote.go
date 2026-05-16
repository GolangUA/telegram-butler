package memory

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/GolangUA/telegram-butler/internal/entity"
)

type VoteRepository struct {
	mu    sync.Mutex
	votes map[string]*entity.Vote
}

func NewVoteRepository() *VoteRepository {
	return &VoteRepository{
		votes: make(map[string]*entity.Vote),
	}
}

func key(chatID, targetUserID int64) string {
	return fmt.Sprintf("%d_%d", chatID, targetUserID)
}

func (r *VoteRepository) Create(_ context.Context, vote *entity.Vote) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.votes[key(vote.ChatID, vote.TargetUserID)] = vote

	return nil
}

func (r *VoteRepository) GetActive(_ context.Context, chatID, targetUserID int64) (*entity.Vote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	vote, ok := r.votes[key(chatID, targetUserID)]
	if !ok || vote.Status != entity.VoteStatusActive {
		return nil, entity.ErrVoteNotFound
	}

	return vote, nil
}

func (r *VoteRepository) AddVoter(
	_ context.Context, chatID, targetUserID int64, voter entity.Voter,
) (*entity.Vote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	vote, ok := r.votes[key(chatID, targetUserID)]
	if !ok || vote.Status != entity.VoteStatusActive {
		return nil, entity.ErrVoteNotFound
	}

	if !slices.ContainsFunc(vote.Voters, func(v entity.Voter) bool { return v.ID == voter.ID }) {
		vote.Voters = append(vote.Voters, voter)
	}

	return vote, nil
}

func (r *VoteRepository) SetStatus(_ context.Context, chatID, targetUserID int64, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	vote, ok := r.votes[key(chatID, targetUserID)]
	if !ok {
		return entity.ErrVoteNotFound
	}

	vote.Status = status

	return nil
}

func (r *VoteRepository) ListActive(_ context.Context) ([]*entity.Vote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	active := make([]*entity.Vote, 0)

	for _, v := range r.votes {
		if v.Status == entity.VoteStatusActive {
			active = append(active, v)
		}
	}

	return active, nil
}
