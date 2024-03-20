package callbackdata

import "fmt"

const (
	AgreeDecision   = "agree"
	DeclineDecision = "decline"
)

func NewAgreeWithGroupID(groupdID int64) string {
	return fmt.Sprintf("%s_%v", AgreeDecision, groupdID)
}

func NewDeclineWithGroupID(groupdID int64) string {
	return fmt.Sprintf("%s_%v", DeclineDecision, groupdID)
}
