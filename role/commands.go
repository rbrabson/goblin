package role

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"guild-admin": guildAdmin,
	}

	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "guild-admin",
			Description: "Commands used to configure the bot for a given server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "role",
					Description: "Manages the admin roles for the bot for this server.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "list",
							Description: "Returns the list of admin roles for the server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
						{
							Name:        "add",
							Description: "Adds an admin role for this server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionRole,
									Name:        "role",
									Description: "The role to add.",
									Required:    true,
								},
							},
						},
						{
							Name:        "remove",
							Description: "Removes an admin role for this server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionRole,
									Name:        "role",
									Description: "The role to remove.",
									Required:    true,
								},
							},
						},
					},
				},
			},
		},
	}
)

// guildAdmin handles the guildAdmin command.
func guildAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if discord.IsShuttingDown(s, i) {
		return
	}

	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have permission to use this command."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response", "error", err)
		}
		return
	}

	options := i.ApplicationCommandData().Options
	if options[0].Name == "role" {
		role(s, i)
	} else {
		slog.Warn("unknown guild-admin command", "guildID", i.GuildID, "userID", i.Member.User.ID, "command", options[0].Name)
	}
}

// role handles the role subcommands for the server command.
func role(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options[0].Options
	switch options[0].Name {
	case "add":
		addRole(s, i)
	case "list":
		listRoles(s, i)
	case "remove":
		removeRole(s, i)
	default:
		slog.Warn("unknown guild-admin role command", "guildID", i.GuildID, "userID", i.Member.User.ID, "command", options[0].Name)
	}
}

// addRole adds a role to the list of admin roles for the server.
func addRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	role := options[0].RoleValue(s, guildID)

	if role == nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The role was not found. Please try again."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response", "error", err)
		}
		return
	}

	guild.SyncGuildRoles(s, guildID)

	// Get the server configuration
	server := guild.GetGuild(guildID)

	// Add the role to the server configuration
	server.AddAdminRole(role.ID)
	slog.Debug("/guild-admin role add", "guildID", guildID, "roleID", role.ID, "roleName", role.Name)

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Role \"%s\" added", role.Name)),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("error sending response", "error", err)
	}
}

// removeRole removes a role from the list of admin roles for the server.
func removeRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	role := options[0].RoleValue(s, guildID)

	if role == nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The role was not found. Please try again."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response", "error", err)
		}
		return
	}

	guild.SyncGuildRoles(s, guildID)

	// Get the server configuration
	server := guild.GetGuild(guildID)

	// Remove the role from the server configuration
	server.RemoveAdminRole(role.ID)
	slog.Debug("/guild-admin role remove", "guildID", guildID, "roleID", role.ID, "roleName", role.Name)

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Role \"%s\" removed", role.Name)),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("error sending response", "error", err)
	}
}

// listRoles lists the admin roles for the server.
func listRoles(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID

	guild.SyncGuildRoles(s, guildID)

	// Get the server configuration
	server := guild.GetGuild(guildID)

	// Get the list of admin roles
	roles := server.GetAdminRoles()

	// Send the list of admin roles to the user
	var sb strings.Builder
	sb.WriteString("**Admin Roles**:\n")
	for _, role := range roles {
		sb.WriteString(role + "\n")
	}
	roleList := sb.String()
	slog.Debug("/guild-admin role list", "guildID", guildID, "roles", len(roles))

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(roleList),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending response", "error", err)
	}
}
