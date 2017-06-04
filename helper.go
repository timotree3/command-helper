// Package commandhelper provides tools for creating command-based Discord bots
package commandhelper

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

func HelpAction(bot Bot, args []string, s *discordgo.Session, m *discordgo.Message) error {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, bot.Help())
	} else {
		s.ChannelMessageSend(m.ChannelID, bot.Usage(args[0]))
	}
	return nil
}

var (
	ErrNotInitialized = errors.New("bot not yet initialized, use AddHandler on Bot.ReadyHandler")
	// HelpCommand the prints the help menu
	//
	// It is bundled by default by NewBot().
	HelpCommand = Command{
		Usage:       "[command]",
		Description: "prints the help menu or given command's usage",
		Action:      HelpAction,
	}
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

// CommandAction is the type of a Command's Action
type Action func(Bot, []string, *discordgo.Session, *discordgo.Message) error

// Command is a sub-command for a Bot.
type Command struct {
	Usage       string       // Available postional arguments shown in help menu
	Description string       // Description show in help menu
	Flags       flag.FlagSet // Custom flags for the command
	Action      Action       // Function to run if command is called
}

// NewBot returns a Bot with bundles HelpCommand.
func NewBot() *Bot {
	return &Bot{
		Commands: map[string]Command{
			"help": HelpCommand,
		},
	}
}

// ReadyHandler gives the bot data made available after opening the discordgo.Session.
//
// Feed this into AddHandler after configuring your bot.
func (bot *Bot) ReadyHandler(s *discordgo.Session, r *discordgo.Ready) {
	bot.prefix = fmt.Sprintf("@%s ", s.State.User.Username)

	// Configure each command's flag system properly
	for name, command := range bot.Commands {
		command.Flags.Init(bot.prefix+" "+name, flag.ContinueOnError)
	}
	bot.initialized = true
}

// MessageHandler reads the configuration of bot to determine how to parse input.
// Feed this into AddHandler after configuring your bot.
func (bot *Bot) MessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if !bot.initialized {
		panic(ErrNotInitialized)
	}

	// If the message was sent by ourselves, ignore it.
	if s.State.User.ID == m.Author.ID {
		return
	}

	argv, err := findArgs(m.Content, []string{s.State.User.Mention()})
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}
	if argv == nil {
		return
	}

	command, ok := bot.Commands[argv[0]]
	if !ok {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("command not found: `%s`", argv[0]))
		return
	}

	command.Flags.Parse(argv[1:])
	err = command.Action(*bot, command.Flags.Args(), s, m.Message)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}
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

OPTIONS:
%s`+"```",
		command.Description,
		bot.prefix+commandName, command.Usage,
		help(command.Flags))
}

func findArgs(text string, prefixes []string) ([]string, error) {
	var fields []string
	for _, prefix := range prefixes {
		if trimmed := strings.TrimPrefix(text, prefix); trimmed != text {
			// Something was removed, so it must begin with that prefix.
			fields = strings.Fields(trimmed)
			break
		}
	}
	if fields == nil {
		return nil, nil
	}

	if len(fields) == 0 {
		return nil, errors.New("no command specified")
	}
	return groupArgs(fields)
}

// Allow arguments to contain spaces if surrounded by backquotes
func groupArgs(fields []string) (argv []string, err error) {
	// Start outside of a group
	grouped := false
	group := ""
	// Keep the command name
	argv = []string{fields[0]}
	for _, arg := range fields[1:] {
		if !grouped && arg != strings.TrimPrefix(arg, "`") {
			grouped = true
			arg = strings.TrimPrefix(arg, "`")
		}
		if grouped && arg != strings.TrimSuffix(arg, "`") {
			grouped = false
			arg = strings.TrimSuffix(arg, "`")
		}
		group += arg
		if grouped {
			// There was a space here before it got split into fields.
			group += " "
		} else {
			argv = append(argv, group)
			group = ""
		}
	}
	if grouped {
		return nil, errors.New("mismatched backquotes ``(`)``")
	}
	return
}

func help(flags flag.FlagSet) string {
	// Create a buffer for the flag's help to output to.
	buf := bytes.NewBuffer(nil)
	flags.SetOutput(buf)
	flags.PrintDefaults()
	// If there are any defined flags...
	if len(buf.String()) > 0 {
		return buf.String()
	} else {
		return "  No options defined."
	}
}
