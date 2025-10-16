package callbackdata

import "fmt"

const (
	AgreeDecision   = "agree"
	DeclineDecision = "decline"
)

func NewAgreeWithGroupID(groupID int64, msgID int) string {
	return fmt.Sprintf("%s_%d_%d", AgreeDecision, groupID, msgID)
}

func NewDeclineWithGroupID(groupID int64, msgID int) string {
	return fmt.Sprintf("%s_%d_%d", DeclineDecision, groupID, msgID)
}
