package shop

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/internal/discmsg"
	log "github.com/sirupsen/logrus"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"show-admin": shopAdmin,
		"shop":       shop,
	}
)

var (
	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "shop-admin",
			Description: "Commands used to interact with the shop for this server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "add",
					Description: "Adds an item to the shop that may be purchased by a member.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "role",
							Description: "Adds a purchasable role to the shop.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "name",
									Description: "The name of the role that may be purchased.",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "description",
									Description: "A brief description of the role to be purchased. This defaults to the role name",
									Required:    false,
								},
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "cost",
									Description: "The cost of the role.",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "duration",
									Description: "The length the role is assigned before being automatically removed.",
									Required:    false,
								},
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "renewable",
									Description: "Whether the member can auto-renew the role purchase when the duration expires. This has no effect unless duration is set to a value.",
									Required:    false,
								},
							},
						},
					},
				},
				{
					Name:        "update",
					Description: "Updates an item in the shop that may be purchased by a member.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "role",
							Description: "Updates a purchasable role to the shop.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "name",
									Description: "The name of the role that may be purchased.",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "description",
									Description: "A brief description of the role to be purchased.",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "cost",
									Description: "The cost of the role.",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "duration",
									Description: "The length the role is assigned before being automatically removed.",
									Required:    false,
								},
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "renewable",
									Description: "Whether the member can auto-renew the role purchase when the duration expires.",
									Required:    false,
								},
							},
						},
					},
				},
				{
					Name:        "remove",
					Description: "Removes an item from the shop.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "role",
							Description: "A role in the shop that is to be removed.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "name",
									Description: "The name of the role to remove from the shop.",
									Required:    true,
								},
							},
						},
					},
				},
				{
					Name:        "list",
					Description: "Lists the items in the shop.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				},
			},
		},
	}

	memberCommands = []*discordgo.ApplicationCommand{
		{
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "buy",
					Description: "Adds an item to the shop that may be purchased by a member.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "role",
							Description: "Buys a custom role for the server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "name",
									Description: "The name of the role to be purchased.",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "renew",
									Description: "Controls whether the role is auto-renewed when the purchase duration expires.",
									Required:    false,
								},
							},
						},
					},
				},
				{
					Name:        "update",
					Description: "Changes the auto-renew status for a role that was previously purchased.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "role",
							Description: "Adds a purchasable role to the shop.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "name",
									Description: "The name of the role to be purchased.",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "renew",
									Description: "Controls whether the role is auto-renewed when the purchase duration expires.",
									Required:    false,
								},
							},
						},
					},
				},
				{
					Name:        "list",
					Description: "Lists the items in the shop that may be purchased.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				},
			},
		},
	}
)

// shopAdmin routes the shop admin commands to the proper handers.
func shopAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.shopAdmin")
	defer log.Trace("<-- shop.shopAdmin")

	log.WithFields(log.Fields{"guildID": i.Member.GuildID, "memberID": i.Member.User.ID}).Warn("TODO: shop-admin not implemented")
	discmsg.SendEphemeralResponse(s, i, "This command is not yet implemented.")
}

// shop routes the shop commands to the proper handers.
func shop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.shop")
	defer log.Trace("<-- shop.shop")

	log.WithFields(log.Fields{"guildID": i.Member.GuildID, "memberID": i.Member.User.ID}).Warn("TODO: shop not implemented")
	discmsg.SendEphemeralResponse(s, i, "This command is not yet implemented.")
}
