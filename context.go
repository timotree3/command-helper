// Package commandhelper provides tools for creating command-based Discord bots
package commandhelper

import (
	"github.com/bwmarrin/discordgo"
)

type Context struct {
	Session     *discordgo.Session
	Message     *discordgo.Message
	commandLine string // The raw text used on the command-line to call the bot.
}
