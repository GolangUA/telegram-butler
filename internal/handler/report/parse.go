package report

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"

	"github.com/GolangUA/telegram-butler/internal/entity"
)

// encodeCallbackData builds the inline-button payload that telego routes back
// to handleVote via the CallbackDataPrefix predicate.
func encodeCallbackData(chatID, targetUserID int64) string {
	return fmt.Sprintf("%s%d_%d", callbackPrefix, chatID, targetUserID)
}

// parseCallbackData is the inverse of encodeCallbackData. The prefix is
// guaranteed by the telego predicate, but we re-check defensively.
func parseCallbackData(data string) (chatID, targetUserID int64, err error) {
	if !strings.HasPrefix(data, callbackPrefix) {
		return 0, 0, fmt.Errorf("missing prefix %q", callbackPrefix)
	}

	rest := strings.TrimPrefix(data, callbackPrefix)

	parts := strings.Split(rest, "_")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected 2 parts, got %d", len(parts))
	}

	chatID, err = parseInt64(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid chatID: %w", err)
	}

	targetUserID, err = parseInt64(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid targetUserID: %w", err)
	}

	return chatID, targetUserID, nil
}

func parseInt64(s string) (int64, error) {
	var n int64

	_, err := fmt.Sscanf(s, "%d", &n)

	return n, err
}

// voterFromUser projects a Telegram user into the persistence-layer Voter.
func voterFromUser(u *telego.User) entity.Voter {
	return entity.Voter{
		ID:        u.ID,
		FirstName: u.FirstName,
		Username:  u.Username,
	}
}
