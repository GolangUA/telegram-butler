package callbackdata

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Payload struct {
	Decision  string
	GroupID   int64
	MessageID int
}

var (
	ErrInvalidCallbackData = errors.New("invalid callback query data token")
	ErrInvalidGroupID      = errors.New("invalid groupID in callback query data")
	ErrInvalidMessageID    = errors.New("invalid messageID in callback query data")
	ErrInvalidDecision     = errors.New("invalid terms of use decision")
)

func Parse(data string) (*Payload, error) {
	splits := strings.Split(data, "_")
	if len(splits) != 3 { //nolint:gomnd,mnd
		return nil, fmt.Errorf("%w: %v", ErrInvalidCallbackData, splits)
	}

	groupID, err := strconv.ParseInt(splits[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidGroupID, splits[1])
	}

	messageID, err := strconv.Atoi(splits[2])
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidMessageID, splits[1])
	}

	decision := splits[0]
	if decision != AgreeDecision && decision != DeclineDecision {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDecision, decision)
	}

	return &Payload{
		Decision:  decision,
		GroupID:   groupID,
		MessageID: messageID,
	}, nil
}
