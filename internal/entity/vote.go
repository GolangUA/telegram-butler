package entity

import (
	"errors"
	"time"
)

const (
	VoteStatusActive  = "active"
	VoteStatusMuted   = "muted"
	VoteStatusExpired = "expired"
)

var ErrVoteNotFound = errors.New("vote not found")

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
