package commands

import (
	"slices"

	"github.com/mymmrac/telego"
)

const (
	SendRules      = "rules"
	SendUsefulInfo = "useful"
	SendHelp       = "help"
	Mute           = "m"
	MuteFull       = "mute"
)

var publicCommands = []telego.BotCommand{
	{Command: SendRules, Description: "правила спільноти"},
	{Command: SendHelp, Description: "інформація про бота"},
	{Command: SendUsefulInfo, Description: "корисна інформація по Go"},
}

// Admins see only the most-specific scope's command list (Telegram does not merge scopes),
// so the admin menu must include the public commands too.
var adminCommands = slices.Concat(publicCommands, []telego.BotCommand{
	{Command: Mute, Description: "mute користувача (reply)"},
	{Command: MuteFull, Description: "mute користувача (reply)"},
})
