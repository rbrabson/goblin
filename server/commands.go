package server

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/internal/discmsg"
	discrole "github.com/rbrabson/goblin/internal/role"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"server-admin": serverAdmin,
	}

	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "server-admin",
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

// serverAdmin handles the serverAdmin command.
func serverAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> server.server")
	defer log.Trace("<-- server.server")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	if !discrole.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := p.Sprintf("You do not have permission to use this command.")
		discmsg.SendEphemeralResponse(s, i, resp)
		return
	}

	options := i.ApplicationCommandData().Options
	if options[0].Name == "role" {
		role(s, i)
	} else {
		log.WithFields(log.Fields{"command": options[0].Name}).Warn("unknown server-admin command")
	}
}

// role handles the role subcommands for the server command.
func role(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> server.role")
	defer log.Trace("<-- server.role")

	options := i.ApplicationCommandData().Options[0].Options
	switch options[0].Name {
	case "add":
		addRole(s, i)
	case "list":
		listRoles(s, i)
	case "remove":
		removeRole(s, i)
	default:
		log.WithFields(log.Fields{"subcommand": options[0].Name}).Warn("unknown server-admin role command")
	}
}

// addRole adds a role to the list of admin roles for the server.
func addRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> server.addRole")
	defer log.Trace("<-- server.addRole")

	guildID := i.GuildID
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	roleName := options[0].StringValue()

	// Get the server configuration
	server := discrole.GetServer(guildID)

	// Add the role to the server configuration
	server.AddAdminRole(roleName)
	log.WithFields(log.Fields{"guild": guildID, "role": roleName}).Debug("/server-admin role add")

	discmsg.SendResponse(s, i, fmt.Sprintf("Role \"%s\" added", roleName))
}

// removeRole removes a role from the list of admin roles for the server.
func removeRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> server.removeRole")
	defer log.Trace("<-- server.removeRole")

	guildID := i.GuildID
	options := i.ApplicationCommandData().Options[0].Options
	roleName := options[0].StringValue()

	// Get the server configuration
	server := discrole.GetServer(guildID)

	// Remove the role from the server configuration
	server.RemoveAdminRole(roleName)
	log.WithFields(log.Fields{"guild": guildID, "role": roleName}).Debug("/server-admin role remove")

	discmsg.SendResponse(s, i, fmt.Sprintf("Role \"%s\" removed", roleName))
}

// listRoles lists the admin roles for the server.
func listRoles(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> server.listRoles")
	defer log.Trace("<-- server.listRoles")

	guildID := i.GuildID

	// Get the server configuration
	server := discrole.GetServer(guildID)

	// Get the list of admin roles
	roles := server.GetAdminRoles()

	// Send the list of admin roles to the user
	var sb strings.Builder
	sb.WriteString("**Admin Roles**:\n")
	for _, role := range roles {
		sb.WriteString(role + "\n")
	}
	roleList := sb.String()
	log.WithFields(log.Fields{"guild": guildID, "roles": roleList}).Debug("/server-admin role list")

	discmsg.SendEphemeralResponse(s, i, roleList)
}
