package duration

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var regex = regexp.MustCompile(`^(\d+)(m|min|minute|h|d|w|mo|month)$`)

// expectedMatches is the number of submatches: full match + number + unit.
const expectedMatches = 3

// Parse parses a duration string like "1h", "30m", "2d", "1w", "1mo".
func Parse(s string) (time.Duration, error) {
	matches := regex.FindStringSubmatch(s)
	if len(matches) != expectedMatches {
		return 0, fmt.Errorf("invalid duration %q,"+
			" expected: <number><m|min|minute|h|d|w|mo|month>"+
			" (e.g. 30m, 1h, 2d, 1w, 1mo)", s)
	}

	n, err := strconv.Atoi(matches[1])
	if err != nil || n <= 0 {
		return 0, errors.New("duration must be a positive number")
	}

	switch matches[2] {
	case "m", "min", "minute":
		return time.Duration(n) * time.Minute, nil
	case "h":
		return time.Duration(n) * time.Hour, nil
	case "d":
		return time.Duration(n) * 24 * time.Hour, nil
	case "w":
		return time.Duration(n) * 7 * 24 * time.Hour, nil
	case "mo", "month":
		return time.Duration(n) * 30 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown duration unit %q", matches[2])
	}
}

// Format formats a duration into a human-readable short form like "1h", "2d", "1w", "1mo".
func Format(d time.Duration) string {
	switch {
	case d%(30*24*time.Hour) == 0:
		months := int(d / (30 * 24 * time.Hour))
		return fmt.Sprintf("%dmo", months)
	case d%(7*24*time.Hour) == 0:
		weeks := int(d / (7 * 24 * time.Hour))
		return fmt.Sprintf("%dw", weeks)
	case d%(24*time.Hour) == 0:
		days := int(d / (24 * time.Hour))
		return fmt.Sprintf("%dd", days)
	case d%time.Hour == 0:
		hours := int(d / time.Hour)
		return fmt.Sprintf("%dh", hours)
	default:
		minutes := int(d / time.Minute)
		return fmt.Sprintf("%dm", minutes)
	}
}
