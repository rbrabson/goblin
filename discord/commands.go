package discord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/dgame/discmsg"
	log "github.com/sirupsen/logrus"
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

	discmsg.SendResponse(s, i, getAdminHelp())
}

// version shows the version of bot you are running.
func version(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> version")
	defer log.Trace("<-- version")

	discmsg.SendEphemeralResponse(s, i, "You are running Heist version "+botVersion+".")
}

// getHelp gets help about commands from all plugins.
func getHelp() string {
	log.Trace("--> discord.getMemberHelp")
	log.Trace("<-- discord.getMemberHelp")

	var sb strings.Builder
	for _, plugin := range ListPlugin() {
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
