package discord

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/rbrabson/disgomsg"
	page "github.com/rbrabson/disgopage"
	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/internal/unicode"
)

const (
	HelpMessagesPerPage  = 5
	HelpPaginatorTimeout = 2 * 60 // 2 minutes
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
									Type:        discordgo.ApplicationCommandOptionUser,
									Name:        "user",
									Description: "The member to add as an owner.",
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
									Type:        discordgo.ApplicationCommandOptionUser,
									Name:        "user",
									Description: "The member to remove as an owner.",
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
									Name:        "member",
									Description: "The member to add as an admin.",
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
									Type:        discordgo.ApplicationCommandOptionUser,
									Name:        "user",
									Description: "The member to remove as an admin.",
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
	sendHelpMessages(s, i, "Help Messages", getHelp())
}

// adminHelp sends a help message for administrative commands.
func adminHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have permission to use this command."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
		return
	}

	sendHelpMessages(s, i, "Admin Help Messages", getAdminHelp())
}

// sendHelpMessages sends help messages in a paginated format.
func sendHelpMessages(s *discordgo.Session, i *discordgo.InteractionCreate, title string, helpMessages [][]string) {
	embedFields := make([]*discordgo.MessageEmbedField, 0, len(helpMessages))

	for _, msg := range helpMessages {
		if len(msg) == 0 {
			continue
		}
		slog.Error("******",
			slog.Any("msg", msg),
		)
		name := msg[0]
		var value string
		if len(msg) > 1 {
			value = strings.Join(msg[1:], "\n")
			value = strings.ReplaceAll(value, "\n\n", "\n")
		}
		embedFields = append(embedFields, &discordgo.MessageEmbedField{
			Name:   name,
			Value:  value,
			Inline: false,
		})
	}

	paginator := page.NewPaginator(
		page.WithDiscordConfig(
			page.DiscordConfig{
				Session:                bot.Session,
				AddComponentHandler:    bot.AddComponentHandler,
				RemoveComponentHandler: bot.RemoveComponentHandler,
			},
		),
		page.WithItemsPerPage(HelpMessagesPerPage),
		page.WithIdleWait(HelpPaginatorTimeout),
	)
	err := paginator.CreateInteractionResponse(s, i, title, embedFields, true)
	if err != nil {
		slog.Error("unable to send admin help commands",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
		return
	}
}

// version shows the version of bot you are running.
func version(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have permission to use this command."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send help response",
				slog.Any("error", err),
			)
		}
		return
	}

	if Revision == "" {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You are running " + BotName + " version " + Version + "."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send help response",
				slog.Any("error", err),
			)
		}
	} else {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You are running " + BotName + " version " + Version + "." + Revision + "."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send help response",
				slog.Any("error", err),
			)
		}
	}
}

// getHelp gets help about commands from all plugins.
func getHelp() [][]string {
	helpMessages := make([][]string, 0, 10)
	for _, plugin := range ListPlugin() {
		pluginMessages := make([]string, 0, 10)
		for _, str := range plugin.GetHelp() {
			str := strings.ReplaceAll(str, "#", "")
			str = strings.TrimSpace(str)
			strs := strings.Split(str, "\n")
			pluginMessages = append(pluginMessages, strs...)
		}
		helpMessages = append(helpMessages, pluginMessages)
	}

	return helpMessages
}

// getAdminHelp returns help about administrative commands for all bots.
func getAdminHelp() [][]string {
	helpMessages := make([][]string, 0, 10)
	for _, plugin := range ListPlugin() {
		pluginMessages := make([]string, 0, 10)
		for _, str := range plugin.GetAdminHelp() {
			str := strings.ReplaceAll(str, "#", "")
			str = strings.TrimSpace(str)
			strs := strings.Split(str, "\n")
			pluginMessages = append(pluginMessages, strs...)
		}
		helpMessages = append(helpMessages, pluginMessages)
	}

	return helpMessages
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
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send help response",
				slog.Any("error", err),
			)
		}
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
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("failed to send response",
			slog.Any("error", err),
		)
	}
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
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
	}

	option := i.ApplicationCommandData().Options[0].Options[0]
	switch option.Name {
	case "add":
		var memberID string
		member, err := guild.GetMemberByUser(s, i.GuildID, option.UserValue(s))
		if err != nil {
			resp := disgomsg.NewResponse(
				disgomsg.WithContent("The user to get the account for was not found. Please try again."),
			)
			if err := resp.SendEphemeral(s, i.Interaction); err != nil {
				slog.Error("error sending response",
					slog.String("guildID", i.GuildID),
					slog.String("error", err.Error()),
				)
			}
			return
		}
		memberID = member.MemberID

		err = server.AddOwner(memberID)
		if err != nil {
			resp := disgomsg.NewResponse(
				disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
			)
			if err := resp.SendEphemeral(s, i.Interaction); err != nil {
				slog.Error("failed to send response",
					slog.Any("error", err),
				)
			}
			slog.Error("failed to add owner",
				slog.String("userID", memberID),
				slog.Any("error", err),
			)
			return
		}
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Added " + memberID + " as a a server owner."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
		slog.Info("added owner",
			slog.String("userID", memberID),
		)
	case "list":
		owers := server.ListOwners()
		if len(owers) == 0 {
			resp := disgomsg.NewResponse(
				disgomsg.WithContent("There are no owners for this server."),
			)
			if err := resp.SendEphemeral(s, i.Interaction); err != nil {
				slog.Error("failed to send response",
					slog.Any("error", err),
				)
			}
			return
		}
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Owners: " + strings.Join(owers, ", ")),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
	case "remove":
		option := i.ApplicationCommandData().Options[0].Options[0].Options[0]
		member, err := guild.GetMemberByUser(s, i.GuildID, option.UserValue(s))
		if err != nil {
			resp := disgomsg.NewResponse(
				disgomsg.WithContent("The requested user cannot be found."),
			)
			if err := resp.SendEphemeral(s, i.Interaction); err != nil {
				slog.Error("error sending response",
					slog.String("guildID", i.GuildID),
					slog.String("error", err.Error()),
				)
			}
			return
		}
		memberID := member.MemberID
		err = server.RemoveOwner(memberID)
		if err != nil {
			resp := disgomsg.NewResponse(
				disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
			)
			if err := resp.SendEphemeral(s, i.Interaction); err != nil {
				slog.Error("failed to send response",
					slog.Any("error", err),
				)
			}
			slog.Error("failed to remove owner",
				slog.String("userID", memberID),
				slog.Any("error", err),
			)
			return
		}
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Removed " + memberID + " as a server owner."),
		)
		if err := resp.Send(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
		slog.Info("removed owner",
			slog.String("userID", memberID),
		)
	default:
		slog.Error("unknown subcommand",
			slog.String("subCommand", option.Name),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Unknown subcommand: " + option.Name),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
	}
}

// manageAdmins manages the server admins.
func manageAdmins(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := GetServer()

	options := i.ApplicationCommandData().Options[0].Options
	switch options[0].Name {
	case "add":
		option := i.ApplicationCommandData().Options[0].Options[0].Options[0]
		member, err := guild.GetMemberByUser(s, i.GuildID, option.UserValue(s))
		if err != nil {
			resp := disgomsg.NewResponse(
				disgomsg.WithContent("The requested user cannot be found."),
			)
			if err := resp.SendEphemeral(s, i.Interaction); err != nil {
				slog.Error("error sending response",
					slog.String("guildID", i.GuildID),
					slog.String("error", err.Error()),
				)
			}
			return
		}
		memberID := member.MemberID
		err = server.AddAdmin(memberID)
		if err != nil {
			resp := disgomsg.NewResponse(
				disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
			)
			if err := resp.SendEphemeral(s, i.Interaction); err != nil {
				slog.Error("failed to send response",
					slog.Any("error", err),
				)
			}
			slog.Error("failed to add admin",
				slog.String("userID", memberID),
				slog.Any("error", err),
			)
			return
		}
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Added " + memberID + " as a server admin."),
		)
		if err := resp.Send(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
		slog.Error("added admin",
			slog.String("userID", memberID),
		)
	case "list":
		owers := server.ListAdmins()
		if len(owers) == 0 {
			resp := disgomsg.NewResponse(
				disgomsg.WithContent("There are no admins for this server."),
			)
			if err := resp.SendEphemeral(s, i.Interaction); err != nil {
				slog.Error("failed to send response",
					slog.Any("error", err),
				)
			}
			return
		}
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Admins: " + strings.Join(owers, ", ")),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
	case "remove":
		option := i.ApplicationCommandData().Options[0].Options[0].Options[0]
		member, err := guild.GetMemberByUser(s, i.GuildID, option.UserValue(s))
		if err != nil {
			resp := disgomsg.NewResponse(
				disgomsg.WithContent("The requested user cannot be found."),
			)
			if err := resp.SendEphemeral(s, i.Interaction); err != nil {
				slog.Error("error sending response",
					slog.String("guildID", i.GuildID),
					slog.String("error", err.Error()),
				)
			}
			return
		}
		memberID := member.MemberID
		err = server.RemoveAdmin(memberID)
		if err != nil {
			resp := disgomsg.NewResponse(
				disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
			)
			if err := resp.SendEphemeral(s, i.Interaction); err != nil {
				slog.Error("failed to send response",
					slog.Any("error", err),
				)
			}
			slog.Error("failed to add admin",
				slog.String("userID", memberID),
				slog.Any("error", err),
			)
			return
		}
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Removed " + memberID + " as an server admin."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
		slog.Error("removed admin",
			slog.String("userID", memberID),
		)
	default:
		slog.Error("unknown subcommand",
			slog.String("subCommand", options[0].Name),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Unknown subcommand: " + options[0].Name),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
	}
}
