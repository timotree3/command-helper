// Package commandhelper provides tools for creating command-based Discord bots
package commandhelper

import (
	"fmt"
	"strings"

	"github.com/anmitsu/go-shlex"
	"github.com/bwmarrin/discordgo"
)

// MessageHandler reads the configuration of bot to determine how to parse input.
// Feed this into AddHandler after configuring your bot.
func (bot *Bot) MessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if !bot.initialized {
		bot.prefix = fmt.Sprintf("@%s ", s.State.User.Username)
		bot.initialized = true
	}

	// If the message was sent by ourselves, ignore it.
	if s.State.User.ID == m.Author.ID {
		return
	}

	if !bot.isCommand(m) {
		return
	}

	args, err := bot.args(m.Message)
	if err != nil {
		bot.HelpError(s, m, err)
		return
	}

	command, ok := bot.Commands[word]
	if !ok {
		bot.HelpError(s, m, fmt.Sprintf("command not found: `%s`", word))
		return
	}

	err = command.Action(&Context{Session: s,
		Message:     m.Message,
		commandLine: fields[1:]})
	if err != nil {
		Send(s, m, err.Error())
		return
	}
}

func (c *Context) Args() {

}

func (c *Context) Arg(num uint) {
	return c.Args()[num]
}

func (bot *Bot) isCommand(m *discordgo.Message) bool {
	for _, prefix := range bot.prefixes {
		if strings.HasPrefix(m.Content, prefix) {
			m.Content = strings.TrimPrefix(m.Content, prefix)
			return true
		}
	}
	return false
}

type tokenizer shlex.DefaultTokenizer

func (t tokenizer) IsQuote(r rune) bool {
	return shlex.DefaultTokenizer.IsQuote(r) || r == '`'
}

func (bot *Bot) args(m *discordgo.Message) ([]string, error) {
	l := shlex.NewLexerString(fmt.Sprintf("%s %s", bot.Name, m.Content), true, true)
	l.SetTokenizer(tokenizer{})
	return l.Split()
}
