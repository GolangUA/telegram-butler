package callbackdata

import "fmt"

const (
	AgreeDecision   = "agree"
	DeclineDecision = "decline"
)

func NewAgreeWithGroupID(groupdID int64, msgID int) string {
	return fmt.Sprintf("%s_%d_%d", AgreeDecision, groupdID, msgID)
}

func NewDeclineWithGroupID(groupdID int64, msgID int) string {
	return fmt.Sprintf("%s_%d_%d", DeclineDecision, groupdID, msgID)
}
