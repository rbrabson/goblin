package shop

import (
	"github.com/bwmarrin/discordgo"
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
					Name:        "account",
					Description: "Sets the amount of credits for a given member.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "id",
							Description: "The member ID.",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "amount",
							Description: "The amount to set the account to.",
							Required:    true,
						},
					},
				},
				{
					Name:        "balance",
					Description: "Set the default balance for the bank for the server.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "value",
							Description: "The the default balance for the bank for the server.",
							Required:    true,
						},
					},
				},
				{
					Name:        "name",
					Description: "Set the name of the bank for the server.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "value",
							Description: "The the name of the bank for the server.",
							Required:    true,
						},
					},
				},
				{
					Name:        "currency",
					Description: "Set the currency for the server.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "value",
							Description: "The currency to set for the server.",
							Required:    true,
						},
					},
				},
				{
					Name:        "info",
					Description: "Get information about the banking system configuration.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}

	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "shop",
			Description: "Commands used to interact with the shop for this server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "account",
					Description: "Bank account balance for the member.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
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
}

// shop routes the shop commands to the proper handers.
func shop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.shop")
	defer log.Trace("<-- shop.shop")

	log.WithFields(log.Fields{"guildID": i.Member.GuildID, "memberID": i.Member.User.ID}).Warn("TODO: shop not implemented")
}
