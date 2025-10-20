package blackjack

import "github.com/bwmarrin/discordgo"

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"blackjack": blackjack,
	}

	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "blackjack",
			Description: "Interacts with the blackjack table.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "play",
					Description: "Play the blackjack game.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "stats",
					Description: "Shows a user's stats.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The member or member ID.",
							Required:    false,
						},
					},
				},
			},
		},
	}
)

func blackjack(s *discordgo.Session, i *discordgo.InteractionCreate) {
	subCommand := i.ApplicationCommandData().Options[0].Name

	switch subCommand {
	case "play":
		playBlackjack(s, i)
	case "stats":
		showStats(s, i)
	default:
		// Unknown subcommand
	}
}

func playBlackjack(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Implementation of playing blackjack goes here
}

func showStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Implementation of showing stats goes here
}
