package callback

import (
	"fmt"
	"strconv"
	"strings"
)

type CallbackData struct {
	Decision string
	GroupID  int64
}

func parseDecisionAndGroupID(callbackData string) (*CallbackData, error) {
	splits := strings.Split(callbackData, "_")
	if len(splits) != 2 {
		return nil, fmt.Errorf("invalid callback query data token: %v", splits)
	}

	groupID, err := strconv.ParseInt(splits[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid groupID in callback query data: %s", splits[1])
	}

	decision := splits[0]
	if decision != AgreeDecision && decision != DeclineDecision {
		return nil, fmt.Errorf("invalid callback data for terms of use decision: %v", decision)
	}

	return &CallbackData{decision, groupID}, nil
}
