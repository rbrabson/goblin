package discord

import (
	"fmt"
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
		"server": serverAdmin,
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
			Name:        "server",
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
				{
					Name:        "owner",
					Description: "Manages the server owners.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "add",
							Description: "Adds an owner for this server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "id",
									Description: "The ID of the owner to add.",
									Required:    true,
								},
							},
						},
						{
							Name:        "remove",
							Description: "Removes an owner for this server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "id",
									Description: "The ID of the owner to remove.",
									Required:    true,
								},
							},
						},
						{
							Name:        "list",
							Description: "Lists the owners for this server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
					},
				},
				{
					Name:        "admin",
					Description: "Manages the server admins.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "add",
							Description: "Adds an admin for this server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "id",
									Description: "The ID of the admin to add.",
									Required:    true,
								},
							},
						},
						{
							Name:        "remove",
							Description: "Removes an admin for this server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "id",
									Description: "The ID of the admin to remove.",
									Required:    true,
								},
							},
						},
						{
							Name:        "list",
							Description: "Lists the admins for this server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
					},
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
	slog.Debug("plugins",
		slog.Any("plugins", ListPlugin()),
	)
	for _, plugin := range ListPlugin() {
		slog.Debug("plugin",
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
	server := GetServer()
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.HasOwners() && !server.IsOwner(i.Member.User.ID) && !server.IsAdmin(i.Member.User.ID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have permission to use this command."),
		)
		resp.SendEphemeral(s, i.Interaction)
		slog.Info("user does not have permission",
			slog.String("userID", i.Member.User.ID),
			slog.Bool("hasOwners", server.HasOwners()),
			slog.Bool("isOwner", server.IsOwner(i.Member.User.ID)),
			slog.Bool("isAdmin", server.IsAdmin(i.Member.User.ID)),
		)
		return
	}

	subCommand := i.ApplicationCommandData().Options[0]
	slog.Info(fmt.Sprintf("processing `server/%s` command", subCommand.Name),
		slog.String("userID", i.Member.User.ID),
		slog.Bool("hasOwners", server.HasOwners()),
		slog.Bool("isOwner", server.IsOwner(i.Member.User.ID)),
		slog.Bool("isAdmin", server.IsAdmin(i.Member.User.ID)),
	)
	switch subCommand.Name {
	case "shutdown":
		serverShutdown(s, i)
	case "status":
		serverStatus(s, i)
	case "owner":
		manageOwners(s, i)
	case "admin":
		manageAdmins(s, i)
	default:
		slog.Error("unknown subcommand",
			slog.String("subCommand", subCommand.Name),
		)
	}
}

// serverShutdown prepares the server to be serverShutdown.
func serverShutdown(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: make sure the user is an admin or owner
	slog.Info("*** shutting down all bot services ***",
		slog.String("guildID", i.GuildID),
		slog.String("userID", i.Member.User.ID),
	)
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
		slog.Error("failed to send server status",
			slog.Any("error", err),
		)
		return
	}
	slog.Debug("send server status",
		slog.Any("embeds", embeds),
	)
}

// manageOwners manages the server owners.
func manageOwners(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := GetServer()
	if server.HasOwners() && !server.IsOwner(i.Member.User.ID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have permission to use this command."),
		)
		resp.SendEphemeral(s, i.Interaction)
	}

	options := i.ApplicationCommandData().Options[0].Options
	switch options[0].Name {
	case "add":
		userID := i.ApplicationCommandData().Options[0].Options[0].Options[0].StringValue()
		err := server.AddOwner(userID)
		if err != nil {
			resp := disgomsg.NewResponse(
				disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
			)
			resp.SendEphemeral(s, i.Interaction)
			slog.Error("failed to add owner",
				slog.String("userID", userID),
				slog.Any("error", err),
			)
			return
		}
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Added " + userID + " as a a server owner."),
		)
		resp.Send(s, i.Interaction)
		slog.Error("added owner",
			slog.String("userID", userID),
		)
	case "list":
		owers := server.ListOwners()
		if len(owers) == 0 {
			resp := disgomsg.NewResponse(
				disgomsg.WithContent("There are no owners for this server."),
			)
			resp.SendEphemeral(s, i.Interaction)
			return
		}
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Owners: " + strings.Join(owers, ", ")),
		)
		resp.SendEphemeral(s, i.Interaction)
	case "remove":
		userID := i.ApplicationCommandData().Options[0].Options[0].Options[0].StringValue()
		err := server.RemoveOwner(userID)
		if err != nil {
			resp := disgomsg.NewResponse(
				disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
			)
			resp.SendEphemeral(s, i.Interaction)
			slog.Error("failed to remove owner",
				slog.String("userID", userID),
				slog.Any("error", err),
			)
			return
		}
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Removed " + userID + " as a server owner."),
		)
		resp.Send(s, i.Interaction)
		slog.Error("removed owner",
			slog.String("userID", userID),
		)
	default:
		slog.Error("unknown subcommand",
			slog.String("subCommand", options[0].Name),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Unknown subcommand: " + options[0].Name),
		)
		resp.SendEphemeral(s, i.Interaction)
	}
}

// manageAdmins manages the server admins.
func manageAdmins(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := GetServer()

	options := i.ApplicationCommandData().Options[0].Options
	switch options[0].Name {
	case "add":
		userID := i.ApplicationCommandData().Options[0].Options[0].Options[0].StringValue()
		err := server.AddAdmin(userID)
		if err != nil {
			resp := disgomsg.NewResponse(
				disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
			)
			resp.SendEphemeral(s, i.Interaction)
			slog.Error("failed to add admin",
				slog.String("userID", userID),
				slog.Any("error", err),
			)
			return
		}
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Added " + userID + " as a server admin."),
		)
		resp.Send(s, i.Interaction)
		slog.Error("added admin",
			slog.String("userID", userID),
		)
	case "list":
		owers := server.ListAdmins()
		if len(owers) == 0 {
			resp := disgomsg.NewResponse(
				disgomsg.WithContent("There are no admins for this server."),
			)
			resp.SendEphemeral(s, i.Interaction)
			return
		}
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Admins: " + strings.Join(owers, ", ")),
		)
		resp.SendEphemeral(s, i.Interaction)
	case "remove":
		userID := i.ApplicationCommandData().Options[0].Options[0].Options[0].StringValue()
		err := server.RemoveAdmin(userID)
		if err != nil {
			resp := disgomsg.NewResponse(
				disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
			)
			resp.SendEphemeral(s, i.Interaction)
			slog.Error("failed to add admin",
				slog.String("userID", userID),
				slog.Any("error", err),
			)
			return
		}
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Removed " + userID + " as an server admin."),
		)
		resp.Send(s, i.Interaction)
		slog.Error("removed admin",
			slog.String("userID", userID),
		)
	default:
		slog.Error("unknown subcommand",
			slog.String("subCommand", options[0].Name),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Unknown subcommand: " + options[0].Name),
		)
		resp.SendEphemeral(s, i.Interaction)
	}
}
