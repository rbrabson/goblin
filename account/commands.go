package account

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"account": accountAdmin,
	}

	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "account-admin",
			Description: "Commands used to manage accounts for the server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "alt-id",
					Description: "Manages alt IDs for the server.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "list",
							Description: "Returns the list of alt IDs for the server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionUser,
									Name:        "owner",
									Description: "The owner of the alt account.",
									Required:    false,
								},
							},
						},
						{
							Name:        "add",
							Description: "Adds an alt ID for this server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionUser,
									Name:        "owner",
									Description: "The owner of the alt account.",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionUser,
									Name:        "alt",
									Description: "The alt account to add.",
									Required:    true,
								},
							},
						},
						{
							Name:        "remove",
							Description: "Removes an alt ID for this server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionUser,
									Name:        "owner",
									Description: "The owner of the alt account.",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionUser,
									Name:        "alt",
									Description: "The alt account to remove.",
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

// accountAdmin handles the `/account-admin` command.
func accountAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The system is shutting down."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.Any("error", err),
			)
		}
		return
	}

	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have permission to use this command."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.Any("error", err),
			)
		}
		return
	}

	options := i.ApplicationCommandData().Options
	if options[0].Name == "alt-id" {
		altID(s, i)
	} else {
		slog.Warn("unknown alt-admin command",
			slog.String("guildID", i.GuildID),
			slog.String("userID", i.Member.User.ID),
			slog.String("command", options[0].Name),
		)
	}
}

// altID handles the `/account-admin alt-id` subcommands for the server command.
func altID(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options[0].Options
	switch options[0].Name {
	case "add":
		addAltID(s, i)
	case "list":
		listAltIDs(s, i)
	case "remove":
		removeAltID(s, i)
	default:
		slog.Warn("unknown alt-admin alt-id command",
			slog.String("guildID", i.GuildID),
			slog.String("userID", i.Member.User.ID),
			slog.String("command", options[0].Name),
		)
	}
}

// addAltID handles the `/account-admin alt-id add` command.
func addAltID(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID
	options := i.ApplicationCommandData().Options[0].Options[0].Options

	var altID, ownerID string
	for _, option := range options {
		switch option.Name {
		case "alt":
			altID = option.UserValue(s).ID
		case "owner":
			ownerID = option.UserValue(s).ID
		}
	}

	if IsAltID(guildID, altID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Alt-ID <@%s> already exists", altID)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.Any("error", err),
			)
		}
		return
	}

	alt := newAltID(guildID, ownerID, altID)
	if alt == nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Error creating alt ID."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.Any("error", err),
			)
		}
		return
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Alt-ID added, owner=<@%s>, alt=<@%s>", ownerID, altID)),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.Any("error", err),
		)
	}
}

// removeAltID handles the `/account-admin alt-id remove` command.
func removeAltID(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID
	options := i.ApplicationCommandData().Options[0].Options[0].Options

	var altID, ownerID string
	for _, option := range options {
		switch option.Name {
		case "alt":
			altID = option.UserValue(s).ID
		case "owner":
			ownerID = option.UserValue(s).ID
		}
	}

	if !IsAltID(guildID, altID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Alt-ID <@%s> for owner <@%s> does not exist", altID, ownerID)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.Any("error", err),
			)
		}
		return
	}

	err := DeleteAltID(guildID, altID)
	if err != nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Error deleting alt ID."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.Any("error", err),
			)
		}
		return
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Alt-ID removed, owner=<@%s>, alt=<@%s>", ownerID, altID)),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.Any("error", err),
		)
	}
}

// listAltIDs handles the `/account-admin alt-id list` command.
func listAltIDs(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID
	options := i.ApplicationCommandData().Options[0].Options[0].Options

	var ownerID string
	for _, option := range options {
		if option.Name == "owner" {
			ownerID = option.UserValue(s).ID
		}
	}

	alts := GetAllAltIDsForOwner(guildID, ownerID)
	if len(alts) == 0 {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("No alt IDs found for this server."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.Any("error", err),
			)
		}
		return
	}

	slices.SortFunc(alts, func(a, b *AltID) int {
		if a.OwnerID != b.OwnerID {
			if a.OwnerID < b.OwnerID {
				return -1
			}
			return 1
		}
		if a.AltID < b.AltID {
			return -1
		} else if a.AltID > b.AltID {
			return 1
		}
		return 0
	})

	var builder strings.Builder
	builder.WriteString("Alt-IDs for this server:\n")
	for _, alt := range alts {
		builder.WriteString(fmt.Sprintf("- Owner: <@%s>, Alt: <@%s>\n", alt.OwnerID, alt.AltID))
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(builder.String()),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.Any("error", err),
		)
	}
}
