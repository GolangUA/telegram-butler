package duration

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

const (
	day   = 24 * time.Hour
	week  = 7 * day
	month = 30 * day
)

var durationPattern = regexp.MustCompile(`^(\d+)(m|min|minute|h|d|w|mo|month)$`)

// numDurationGroups is the number of submatches: full match + number + unit.
const numDurationGroups = 3

// Parse parses a duration string like "1h", "30m", "2d", "1w", "1mo".
func Parse(s string) (time.Duration, error) {
	matches := durationPattern.FindStringSubmatch(s)
	if len(matches) != numDurationGroups {
		return 0, fmt.Errorf("invalid duration %q,"+
			" expected: <number><m|min|minute|h|d|w|mo|month>"+
			" (e.g. 30m, 1h, 2d, 1w, 1mo)", s)
	}

	number, unit := matches[1], matches[2]

	n, err := strconv.Atoi(number)
	if err != nil || n <= 0 {
		return 0, errors.New("duration must be a positive number")
	}

	switch unit {
	case "m", "min", "minute":
		return time.Duration(n) * time.Minute, nil
	case "h":
		return time.Duration(n) * time.Hour, nil
	case "d":
		return time.Duration(n) * day, nil
	case "w":
		return time.Duration(n) * week, nil
	case "mo", "month":
		return time.Duration(n) * month, nil
	default:
		return 0, fmt.Errorf("unknown duration unit %q", unit)
	}
}

// Format formats a duration into a human-readable form like "1 hour", "2 days", "1 week", "3 months".
func Format(d time.Duration) string {
	switch {
	case d%month == 0:
		return pluralize(int(d/month), "month")
	case d%week == 0:
		return pluralize(int(d/week), "week")
	case d%day == 0:
		return pluralize(int(d/day), "day")
	case d%time.Hour == 0:
		return pluralize(int(d/time.Hour), "hour")
	default:
		return pluralize(int(d/time.Minute), "minute")
	}
}

func pluralize(n int, unit string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, unit)
	}

	return fmt.Sprintf("%d %ss", n, unit)
}
