package role

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
	log "github.com/sirupsen/logrus"
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
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "name",
									Description: "The name of the role to add.",
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
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "name",
									Description: "The name of the role to remove.",
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
	if status == discord.STOPPING || status == discord.STOPPED {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The system is shutting down."),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have permission to use this command."),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	options := i.ApplicationCommandData().Options
	if options[0].Name == "role" {
		role(s, i)
	} else {
		log.WithFields(log.Fields{"command": options[0].Name}).Warn("unknown guild-admin command")
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
		log.WithFields(log.Fields{"subcommand": options[0].Name}).Warn("unknown guild-admin role command")
	}
}

// addRole adds a role to the list of admin roles for the server.
func addRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	roleName := options[0].StringValue()

	// Get the server configuration
	server := guild.GetGuild(guildID)

	// Add the role to the server configuration
	server.AddAdminRole(roleName)
	log.WithFields(log.Fields{"guild": guildID, "role": roleName}).Debug("/guild-admin role add")

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Role \"%s\" added", roleName)),
	)
	resp.Send(s, i.Interaction)
}

// removeRole removes a role from the list of admin roles for the server.
func removeRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID
	options := i.ApplicationCommandData().Options[0].Options
	roleName := options[0].StringValue()

	// Get the server configuration
	server := guild.GetGuild(guildID)

	// Remove the role from the server configuration
	server.RemoveAdminRole(roleName)
	log.WithFields(log.Fields{"guild": guildID, "role": roleName}).Debug("/guild-admin role remove")

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Role \"%s\" removed", roleName)),
	)
	resp.Send(s, i.Interaction)
}

// listRoles lists the admin roles for the server.
func listRoles(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID

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
	log.WithFields(log.Fields{"guild": guildID, "roles": roleList}).Debug("/guild-admin role list")

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(roleList),
	)
	resp.SendEphemeral(s, i.Interaction)
}
