package discord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"

	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/internal/discmsg"
)

var (
	helpCommandHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"help":      help,
		"adminhelp": adminHelp,
		"version":   version,
	}

	helpCommands = []*discordgo.ApplicationCommand{
		{

			Name:        "help",
			Description: "Provides a description of commands for this server.",
		},
		{
			Name:        "adminhelp",
			Description: "Provides a description of admin commands for this server.",
		},
		{
			Name:        "version",
			Description: "Returns the version of heist running on the server.",
		},
	}
)

// help sends a help message for plugin commands.
func help(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> help")
	log.Trace("<-- help")

	discmsg.SendEphemeralResponse(s, i, getHelp())
}

// adminHelp sends a help message for administrative commands.
func adminHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> adminHelp")
	log.Trace("<-- adminHelp")

	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		p := discmsg.GetPrinter(language.AmericanEnglish)
		resp := p.Sprintf("You do not have permission to use this command.")
		discmsg.SendEphemeralResponse(s, i, resp)
		return
	}

	discmsg.SendEphemeralResponse(s, i, getAdminHelp())
}

// version shows the version of bot you are running.
func version(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> version")
	defer log.Trace("<-- version")

	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		p := discmsg.GetPrinter(language.AmericanEnglish)
		resp := p.Sprintf("You do not have permission to use this command.")
		discmsg.SendEphemeralResponse(s, i, resp)
		return
	}

	discmsg.SendEphemeralResponse(s, i, "You are running "+BotName+" version "+Version+"-"+Revision+".")
}

// getHelp gets help about commands from all plugins.
func getHelp() string {
	log.Trace("--> discord.getMemberHelp")
	log.Trace("<-- discord.getMemberHelp")

	var sb strings.Builder
	log.WithFields(log.Fields{"plugins": ListPlugin()}).Debug("plugins")
	for _, plugin := range ListPlugin() {
		log.WithFields(log.Fields{"plugin": plugin.GetName()}).Debug("plugin")
		for _, str := range plugin.GetHelp() {
			sb.WriteString(str)
		}
	}

	return sb.String()
}

// getAdminHelp returns help about administrative commands for all bots.
func getAdminHelp() string {
	log.Trace("--> discord.getAdminHelp")
	log.Trace("<-- discord.getAdminHelp")

	var sb strings.Builder
	for _, plugin := range ListPlugin() {
		for _, str := range plugin.GetAdminHelp() {
			sb.WriteString(str)
		}
	}

	return sb.String()
}
