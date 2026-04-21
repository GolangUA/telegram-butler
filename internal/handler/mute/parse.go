package mute

import (
	"errors"
	"strings"
	"time"

	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/GolangUA/telegram-butler/internal/duration"
)

type command struct {
	Duration time.Duration
	Reason   string
}

func parseCommand(text string) (*command, error) {
	_, _, args := tu.ParseCommand(text)

	if len(args) == 0 {
		return nil, errors.New("duration is required")
	}

	d, err := duration.Parse(args[0])
	if err != nil {
		return nil, err
	}

	return &command{
		Duration: d,
		Reason:   strings.Join(args[1:], " "),
	}, nil
}
