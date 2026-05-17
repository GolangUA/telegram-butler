package entity

import (
	"errors"
	"fmt"
	"time"
)

const (
	VoteStatusActive  = "active"
	VoteStatusMuted   = "muted"
	VoteStatusExpired = "expired"
)

var ErrVoteNotFound = errors.New("vote not found")

// VoteKey identifies an in-flight vote by (chat, target user) pair.
type VoteKey struct {
	ChatID       int64
	TargetUserID int64
}

func (k VoteKey) String() string {
	return fmt.Sprintf("%d_%d", k.ChatID, k.TargetUserID)
}

type Voter struct {
	ID        int64
	FirstName string
	Username  string
}

type Vote struct {
	ChatID       int64
	TargetUserID int64
	TargetName   string
	ReporterID   int64
	MessageID    int
	ThreadID     int
	Voters       []Voter
	Status       string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}
