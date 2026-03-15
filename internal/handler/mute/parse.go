package mute

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	tu "github.com/mymmrac/telego/telegoutil"
)

var durationRegex = regexp.MustCompile(`^(\d+)(m|min|minute|h|d|w|mo|month)$`)

type command struct {
	Duration time.Duration
	Reason   string
}

func parseCommand(text string) (*command, error) {
	_, _, args := tu.ParseCommand(text)

	if len(args) == 0 {
		return nil, errors.New("duration is required")
	}

	duration, err := parseDuration(args[0])
	if err != nil {
		return nil, err
	}

	cmd := &command{
		Duration: duration,
	}

	if len(args) > 1 {
		cmd.Reason = strings.Join(args[1:], " ")
	}

	return cmd, nil
}

func parseDuration(s string) (time.Duration, error) {
	matches := durationRegex.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("invalid duration %q, expected: <number><m|h|d|w|mo> (e.g. 30m, 1h, 2d, 1w, 1mo)", s)
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
