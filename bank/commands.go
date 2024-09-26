package bank

import "github.com/bwmarrin/discordgo"

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"bank":     bank,
		"account":  account,
		"balance":  currentBalance,
		"monthly":  monthlyBalance,
		"lifetime": lifetimeBalance,
	}

	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "bank",
			Description: "Commands used to interact with the economy for this server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "account",
					Description: "Gets the bank account information for the given member.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "id",
							Description: "The member ID.",
							Required:    true,
						},
					},
				},
				{
					Name:        "set",
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
					Name:        "transfer",
					Description: "Transfers the account balance from one account to another.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "from",
							Description: "The ID of the account to transfer credits from.",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "to",
							Description: "The ID of the account to receive account balance.",
							Required:    true,
						},
					},
				},
				{
					Name:        "channel",
					Description: "Sets the channel ID where the monthly leaderboard is published at the end of the month.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "id",
							Description: "The channel ID.",
							Required:    true,
						},
					},
				},
			},
		},
	}

	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "monthly",
			Description: "Gets the monthly economy leaderboard.",
		},
		{
			Name:        "lifetime",
			Description: "Gets the lifetime economy leaderboard.",
		},
		{
			Name:        "balance",
			Description: "Bank account balance for the member",
		},
	}
)

// bank routes the bank commands to the proper handers.
func bank(s *discordgo.Session, i *discordgo.InteractionCreate) {
}

// account returns information about a bank account for the specified member.
func account(s *discordgo.Session, i *discordgo.InteractionCreate) {
}

// currentBalance returns information about a member's bank account to that member.
func currentBalance(s *discordgo.Session, i *discordgo.InteractionCreate) {
}

// monthlyBalance returns the top 10 monthly players in the server's economy.
func monthlyBalance(s *discordgo.Session, i *discordgo.InteractionCreate) {
}

// monthlyBalance returns the top 10 lifetime players in the server's economy.
func lifetimeBalance(s *discordgo.Session, i *discordgo.InteractionCreate) {
}
