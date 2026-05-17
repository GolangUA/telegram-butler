package entity

import (
	"errors"
	"fmt"
	"time"
)

// VoteStatus is the lifecycle state of a Vote.
type VoteStatus string

const (
	VoteStatusActive  VoteStatus = "active"
	VoteStatusMuted   VoteStatus = "muted"
	VoteStatusExpired VoteStatus = "expired"
)

// ErrVoteNotFound is returned by Repository methods and by the report Service
// when no active vote exists for the given key. Compare with `errors.Is`.
var ErrVoteNotFound = errors.New("vote not found")

// VoteKey identifies an in-flight vote by (chat, target user) pair.
type VoteKey struct {
	ChatID       int64
	TargetUserID int64
}

func (k VoteKey) String() string {
	return fmt.Sprintf("%d_%d", k.ChatID, k.TargetUserID)
}

// Voter is a single participant in a vote.
type Voter struct {
	ID        int64
	FirstName string
	Username  string
}

// Vote is the full state of one community-mute vote, persisted by Repository
// and held in memory by the per-vote coordinator goroutine.
type Vote struct {
	ChatID       int64
	TargetUserID int64
	TargetName   string
	ReporterID   int64
	MessageID    int
	ThreadID     int
	Voters       []Voter
	Status       VoteStatus
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

// Key extracts the identifying (chat, target) pair from the vote.
func (v Vote) Key() VoteKey {
	return VoteKey{ChatID: v.ChatID, TargetUserID: v.TargetUserID}
}
