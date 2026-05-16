package mute

import (
	"errors"
	"strings"
	"time"

	"github.com/mymmrac/telego"
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

// resolveTarget extracts the target user from the replied message. The /mute
// command must be sent as a reply to the user being muted.
func resolveTarget(message telego.Message) (*telego.User, error) {
	if message.ReplyToMessage == nil || message.ReplyToMessage.From == nil {
		return nil, errors.New("command must be a reply to the target user's message")
	}

	return message.ReplyToMessage.From, nil
}
