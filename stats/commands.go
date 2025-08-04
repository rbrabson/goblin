package stats

import (
	"github.com/bwmarrin/discordgo"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		// "stats-admin": statsAdmin,
	}

	adminCommands = []*discordgo.ApplicationCommand{
		// {
		// 	Name:        "stats-admin",
		// 	Description: "Commands used to interact with the stats system.",
		// 	Options: []*discordgo.ApplicationCommandOption{
		// 		{
		// 			Name:        "unique",
		// 			Description: "View player stats",
		// 			Type:        discordgo.ApplicationCommandOptionSubCommand,
		// 			Options: []*discordgo.ApplicationCommandOption{
		// 				{
		// 					Name:        "game",
		// 					Description: "The number of unique players for a game.",
		// 					Type:        discordgo.ApplicationCommandOptionString,
		// 					Required:    true,
		// 					Choices: []*discordgo.ApplicationCommandOptionChoice{
		// 						{
		// 							Name:  "Heist",
		// 							Value: "heist",
		// 						},
		// 						{
		// 							Name:  "Race",
		// 							Value: "race",
		// 						},
		// 					},
		// 				},
		// 				{
		// 					Name:        "type",
		// 					Description: "The type of unique players to view.",
		// 					Type:        discordgo.ApplicationCommandOptionString,
		// 					Required:    true,
		// 					Choices: []*discordgo.ApplicationCommandOptionChoice{
		// 						{
		// 							Name:  "Average",
		// 							Value: "average",
		// 						},
		// 						{
		// 							Name:  "Previous",
		// 							Value: "previous",
		// 						},
		// 					},
		// 				},
		// 				{
		// 					Name:        "period",
		// 					Description: "The time period for the stats.",
		// 					Type:        discordgo.ApplicationCommandOptionString,
		// 					Required:    true,
		// 					Choices: []*discordgo.ApplicationCommandOptionChoice{
		// 						{
		// 							Name:  "Daily",
		// 							Value: "daily",
		// 						},
		// 						{
		// 							Name:  "Weekly",
		// 							Value: "weekly",
		// 						},
		// 						{
		// 							Name:  "Monthly",
		// 							Value: "monthly",
		// 						},
		// 					},
		// 				},
		// 			},
		// 		},
		// 	},
		// },
	}

	memberCommands = []*discordgo.ApplicationCommand{}
)

// statsAdmin handles the /stats-admin command.
func statsAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) {

}
