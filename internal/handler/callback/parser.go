package callback

import (
	"fmt"
	"strconv"
	"strings"
)

func parseDecisionAndGroupID(callbackData string) (string, int64, error) {
	splits := strings.Split(callbackData, "_")
	if len(splits) != 2 {
		return "", 0, fmt.Errorf("invalid callback query data token: %v", splits)
	}

	groupID, err := strconv.ParseInt(splits[1], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid groupID in callback query data: %s", splits[1])
	}

	decision := splits[0]
	if decision != AgreeDecision && decision != DeclineDecision {
		return "", 0, fmt.Errorf("invalid callback data for terms of use decision: %v", decision)
	}

	return decision, groupID, nil
}
