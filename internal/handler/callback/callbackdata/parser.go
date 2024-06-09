package callbackdata

import (
	"fmt"
	"strconv"
	"strings"
)

type Payload struct {
	Decision  string
	GroupID   int64
	MessageID int
}

func Parse(data string) (*Payload, error) {
	splits := strings.Split(data, "_")
	if len(splits) != 3 { //nolint:gomnd,mnd
		return nil, fmt.Errorf("invalid callback query data token: %v", splits)
	}

	groupID, err := strconv.ParseInt(splits[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid groupID in callback query data: %s", splits[1])
	}

	messageID, err := strconv.Atoi(splits[2])
	if err != nil {
		return nil, fmt.Errorf("invalid messageID in callback query data: %s", splits[1])
	}

	decision := splits[0]
	if decision != AgreeDecision && decision != DeclineDecision {
		return nil, fmt.Errorf("invalid callback data for terms of use decision: %v", decision)
	}

	return &Payload{
		Decision:  decision,
		GroupID:   groupID,
		MessageID: messageID,
	}, nil
}
