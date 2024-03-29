package commands

import "github.com/mymmrac/telego"

const (
	SendRules      = "rules"
	SendUsefulInfo = "useful"
	SendHelp       = "help"
)

var allCommands = []telego.BotCommand{
	{Command: SendRules, Description: "правила спільноти"},
	{Command: SendHelp, Description: "інформація про бота"},
	{Command: SendUsefulInfo, Description: "корисна інформація по Go"},
}
