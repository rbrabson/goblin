package discord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"

	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/internal/unicode"
)

var (
	helpCommandHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"help":      help,
		"adminhelp": adminHelp,
		"version":   version,
	}
	serverCommandHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"server-admin": serverAdmin,
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

	serverCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "server-admin",
			Description: "Commands used to interact with the server itself.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "shutdown",
					Description: "Prepares the server to be shutdown.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},

				{
					Name:        "status",
					Description: "Returns the status of the server.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}
)

// help sends a help message for plugin commands.
func help(s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(getHelp()),
		disgomsg.WithInteraction(i.Interaction))
	resp.SendEphemeral(s)
}

// adminHelp sends a help message for administrative commands.
func adminHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have permission to use this command."),
			disgomsg.WithInteraction(i.Interaction),
		)
		resp.SendEphemeral(s)
		return
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(getAdminHelp()),
		disgomsg.WithInteraction(i.Interaction))
	resp.SendEphemeral(s)
}

// version shows the version of bot you are running.
func version(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have permission to use this command."),
			disgomsg.WithInteraction(i.Interaction))
		resp.SendEphemeral(s)
		return
	}

	if Revision == "" {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You are running "+BotName+" version "+Version+"."),
			disgomsg.WithInteraction(i.Interaction),
		)
		resp.SendEphemeral(s)
	} else {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You are running "+BotName+" version "+Version+"."+Revision+"."),
			disgomsg.WithInteraction(i.Interaction))
		resp.SendEphemeral(s)
	}
}

// getHelp gets help about commands from all plugins.
func getHelp() string {
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
	var sb strings.Builder
	for _, plugin := range ListPlugin() {
		for _, str := range plugin.GetAdminHelp() {
			sb.WriteString(str)
		}
	}

	return sb.String()
}

// serverAdmin handles server admin commands.
func serverAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have permission to use this command."),
			disgomsg.WithInteraction(i.Interaction))
		resp.SendEphemeral(s)
		return
	}

	subCommand := i.ApplicationCommandData().Options[0]
	switch subCommand.Name {
	case "shutdown":
		serverShutdown(s, i)
	case "status":
		serverStatus(s, i)
	default:
		log.WithFields(log.Fields{"subCommand": subCommand}).Error("unknown subcommand")
	}
}

// serverShutdown prepares the server to be serverShutdown.
func serverShutdown(s *discordgo.Session, i *discordgo.InteractionCreate) {
	for _, plugin := range ListPlugin() {
		plugin.Stop()
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithContent("Shutting down all bot services."),
		disgomsg.WithInteraction(i.Interaction),
	)
	resp.Send(s)
}

// serverStatus returns the status of the server.
func serverStatus(s *discordgo.Session, i *discordgo.InteractionCreate) {
	plugins := ListPlugin()
	pluginStatus := make([]*discordgo.MessageEmbedField, 0, len(plugins))

	for _, plugin := range plugins {
		pluginStatus = append(pluginStatus, &discordgo.MessageEmbedField{
			Name:   unicode.FirstToUpper(plugin.GetName()),
			Value:  plugin.Status().String(),
			Inline: true,
		})
	}

	embeds := []*discordgo.MessageEmbed{
		{
			Title:  "Plugin Status",
			Fields: pluginStatus,
		},
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("failed to send server status")
		return
	}
	log.WithFields(log.Fields{"embeds": embeds}).Debug("send server status")
}
