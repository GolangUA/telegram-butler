package commands

import "github.com/mymmrac/telego"

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

var adminCommands = []telego.BotCommand{
	{Command: Mute, Description: "мут користувача (reply)"},
	{Command: MuteFull, Description: "мут користувача (reply)"},
}
