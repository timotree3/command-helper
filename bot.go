// Package commandhelper provides tools for creating command-based Discord bots
package commandhelper

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// Bot is the main structure of a discord bot.
// It is recommended to create these using NewBot
type Bot struct {
	Name        string             // Name shown in help menu
	Description string             // Description shown in help menu
	Commands    map[string]Command // Mapping of command name to Command
	prefix      string             // Human-readable @mention tag for help menu
	initialized bool               // Turned on when ReadyHandler is run.
}

// Command is a sub-command for a Bot.
type Command struct {
	Usage       string               // Available postional arguments shown in help menu
	Description string               // Description show in help menu
	Action      func(*Context) error // Function to run if command is called
}

// NewBot returns a Bot with bundles HelpCommand.
func NewBot() *Bot {
	bot := new(Bot)
	bot.Commands = map[string]Command{
		"help": bot.HelpCommand(),
	}
	return bot
}

// HelpCommand returns a command which will print the help menu for it's bot.
//
// It is bundled by default by NewBot().
func (bot *Bot) HelpCommand() Command {
	return Command{
		Usage:       "[command]",
		Description: "prints the help menu or given command's usage",
		Action:      bot.HelpAction,
	}
}

func (bot *Bot) HelpAction(c *Context) error {
	if len(c.Args()) == 0 {
		c.Session.ChannelMessageSend(c.Message.ChannelID, bot.Help())
	} else {
		c.Session.ChannelMessageSend(c.Message.ChannelID, bot.Usage(c.Arg(0)))
	}
	return nil
}

// Help returns the global help menu for a bot
func (bot Bot) Help() string {
	var allCommands string
	for name, command := range bot.Commands {
		allCommands += fmt.Sprintf("\n  %-12s %s", name, command.Description)
	}
	return fmt.Sprintf(`
Showing help for %s:
`+"```"+`
DESCRIPTION:
  %s

USAGE:
  %ssubcommand [arguments]

COMMANDS:%s
`+"```",
		bot.Name,
		bot.Description,
		bot.prefix,
		allCommands)
}

// Usage returns the command specific help menu for a bot.
func (bot Bot) Usage(commandName string) string {
	command, ok := bot.Commands[commandName]
	if !ok {
		return fmt.Sprintf("command not found: `%s`", commandName)
	}
	return fmt.Sprintf(`
Showing usage for subcommand:
`+"```"+`
DESCRIPTION:
  %s

USAGE:
  %s [options] %s
`+"```",
		command.Description,
		bot.prefix+commandName, command.Usage)
}

func Send(s *discordgo.Session, m *discordgo.Message, text string) {
	s.ChannelMessageSend(m.ChannelID, text)
}

func (bot *Bot) HelpError(s *discordgo.Session, m *discordgo.Message, err error) {
	Send(s, m, err.Error()+bot.Help())
}
