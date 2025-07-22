package stats

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/rbrabson/disgomsg"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"stats-admin": statsAdmin,
	}

	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "stats-admin",
			Description: "Commands used to interact with the stats system.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "unique",
					Description: "View your stats",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "game",
							Description: "The number of unique players for a game.",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Heist",
									Value: "heist",
								},
								{
									Name:  "Race",
									Value: "race",
								},
							},
						},
						{
							Name:        "type",
							Description: "The type of unique players to view.",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Average",
									Value: "average",
								},
								{
									Name:  "Total",
									Value: "total",
								},
							},
						},
						{
							Name:        "period",
							Description: "The time period for the stats.",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Daily",
									Value: "daily",
								},
								{
									Name:  "Weekly",
									Value: "weekly",
								},
								{
									Name:  "Monthly",
									Value: "monthly",
								},
								{
									Name:  "Yearly",
									Value: "yearly",
								},
							},
						},
					},
				},
			},
		},
	}

	memberCommands = []*discordgo.ApplicationCommand{}
)

// statsAdmin handles the /stats-admin command.
func statsAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The system is shutting down."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", i.GuildID),
				slog.String("error", err.Error()),
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
				slog.String("guildID", i.GuildID),
				slog.String("error", err.Error()),
			)
		}
		return
	}

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "unique":
		getUniquePlayers(s, i)
	default:
		slog.Warn("unknown bank-admin command",
			slog.String("command", options[0].Name),
		)
	}
}

// getUniquePlayers retrieves the number of unique players for a game. This can be either total
// or average based on the type specified.
func getUniquePlayers(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options[0].Options

	var game, typeValue, period string

	for _, opt := range options {
		switch opt.Name {
		case "game":
			game = opt.StringValue()
		case "type":
			typeValue = opt.StringValue()
		case "period":
			period = opt.StringValue()
		}
	}

	content := p.Sprintf("You requested the %s number of %s unique players for the %s game in guild %s.", typeValue, period, game, i.GuildID)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(content),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("error", err.Error()),
		)
	}
}
