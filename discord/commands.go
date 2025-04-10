package discord

import (
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"

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
	)
	resp.SendEphemeral(s, i.Interaction)
}

// adminHelp sends a help message for administrative commands.
func adminHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have permission to use this command."),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(getAdminHelp()),
	)
	resp.SendEphemeral(s, i.Interaction)
}

// version shows the version of bot you are running.
func version(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have permission to use this command."),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	if Revision == "" {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You are running " + BotName + " version " + Version + "."),
		)
		resp.SendEphemeral(s, i.Interaction)
	} else {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You are running " + BotName + " version " + Version + "." + Revision + "."),
		)
		resp.SendEphemeral(s, i.Interaction)
	}
}

// getHelp gets help about commands from all plugins.
func getHelp() string {
	var sb strings.Builder
	sslog.Debug("plugins",
		"plugins", ListPlugin(),
	)
	for _, plugin := range ListPlugin() {
		sslog.Debug("plugin",
			slog.String("plugin", plugin.GetName()),
		)
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
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	subCommand := i.ApplicationCommandData().Options[0]
	switch subCommand.Name {
	case "shutdown":
		serverShutdown(s, i)
	case "status":
		serverStatus(s, i)
	default:
		sslog.Error("unknown subcommand",
			slog.String("subCommand", subCommand.Name),
		)
	}
}

// serverShutdown prepares the server to be serverShutdown.
func serverShutdown(s *discordgo.Session, i *discordgo.InteractionCreate) {
	for _, plugin := range ListPlugin() {
		plugin.Stop()
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithContent("Shutting down all bot services."),
	)
	resp.Send(s, i.Interaction)
}

// serverStatus returns the status of the server.
func serverStatus(s *discordgo.Session, i *discordgo.InteractionCreate) {
	plugins := ListPlugin()
	pluginStatus := make([]*discordgo.MessageEmbedField, 0, len(plugins))

	botStatus := "Running"
	for _, plugin := range plugins {
		switch plugin.Status() {
		case RUNNING:
			botStatus = "Running"
		case STOPPING:
			botStatus = "Stopping"
		case STOPPED:
			if botStatus != "Stopping" {
				botStatus = "Stopped"
			}
		}
		pluginStatus = append(pluginStatus, &discordgo.MessageEmbedField{
			Name:   unicode.FirstToUpper(plugin.GetName()),
			Value:  plugin.Status().String(),
			Inline: true,
		})
	}

	embeds := []*discordgo.MessageEmbed{
		{
			Title:       "Server Status",
			Description: botStatus,
		},
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
		sslog.Error("failed to send server status",
			slog.String("error", err.Error()),
		)
		return
	}
	sslog.Debug("send server status",
		"embeds", embeds,
	)
}
