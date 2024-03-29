package commands

import "github.com/mymmrac/telego"

const (
	SendRules      = "rules"
	SendUsefulInfo = "useful"
	SendHelp       = "help"
)

var allCommands = []telego.BotCommand{
	{Command: SendRules, Description: "правила спільноти"},
	{Command: SendUsefulInfo, Description: "інформація про бота"},
	{Command: SendHelp, Description: "корисна інформація (курси, документацію, платформи для навчання) по Go"},
}
