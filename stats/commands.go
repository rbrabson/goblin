package stats

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	Heist = "heist"
	Race  = "race"
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
					Name:        "activity",
					Description: "View player activity stats.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "game",
							Description: "The game for which to retrieve the stats.",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Heist",
									Value: Heist,
								},
								{
									Name:  "Race",
									Value: Race,
								},
							},
						},
						{
							Name:        "after",
							Description: "The time period to check the player activity.",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "One Day",
									Value: OneDay,
								},
								{
									Name:  "One Week",
									Value: OneWeek,
								},
								{
									Name:  "Three Months",
									Value: ThreeMonths,
								},
								{
									Name:  "Six Months",
									Value: SixMonths,
								},
								{
									Name:  "Nine Months",
									Value: NineMonths,
								},
								{
									Name:  "Twelve Months",
									Value: TwelveMonths,
								},
							},
						},
						{
							Name:        "since",
							Description: "The time period to check the player activity.",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    false,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Last Week",
									Value: LastWeek,
								},
								{
									Name:  "Last Month",
									Value: LastMonth,
								},
								{
									Name:  "Three Months Ago",
									Value: ThreeMonthsAgo,
								},
								{
									Name:  "Six Months Ago",
									Value: SixMonthsAgo,
								},
								{
									Name:  "Nine Months Ago",
									Value: NineMonthsAgo,
								},
								{
									Name:  "Twelve Months Ago",
									Value: TwelveMonthsAgo,
								},
							},
						},
					},
				},
				{
					Name:        "churn",
					Description: "View player churn.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "game",
							Description: "The game for which to determine the churn.",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Heist",
									Value: Heist,
								},
								{
									Name:  "Race",
									Value: Race,
								},
							},
						},
						{
							Name:        "after",
							Description: "The time period to check the player churn.",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "One Day",
									Value: OneDay,
								},
								{
									Name:  "One Week",
									Value: OneWeek,
								},
								{
									Name:  "Three Months",
									Value: ThreeMonths,
								},
								{
									Name:  "Six Months",
									Value: SixMonths,
								},
								{
									Name:  "Nine Months",
									Value: NineMonths,
								},
								{
									Name:  "Twelve Months",
									Value: TwelveMonths,
								},
							},
						},
						{
							Name:        "since",
							Description: "The time period to check the player churn.",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    false,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Last Week",
									Value: LastWeek,
								},
								{
									Name:  "Last Month",
									Value: LastMonth,
								},
								{
									Name:  "Three Months Ago",
									Value: ThreeMonthsAgo,
								},
								{
									Name:  "Six Months Ago",
									Value: SixMonthsAgo,
								},
								{
									Name:  "Nine Months Ago",
									Value: NineMonthsAgo,
								},
								{
									Name:  "Twelve Months Ago",
									Value: TwelveMonthsAgo,
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
			slog.Error("failed to send response",
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
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
		return
	}

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "activity":
		playerActivity(s, i)
	case "churn":
		playerChurn(s, i)
	}
}

func playerActivity(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	var game, after, since string
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		switch option.Name {
		case "game":
			game = option.StringValue()
		case "after":
			after = option.StringValue()
		case "since":
			since = option.StringValue()
		}
	}

	slog.Debug("Player activity command received",
		slog.String("guild_id", i.GuildID),
		slog.String("game", game),
		slog.String("after", after),
		slog.String("since", since),
	)

	duration := getDuration(after)
	checkAfter := getTime(since)

	// guildID := i.GuildID
	guildID := "236523452230533121"
	activity, err := GetPlayerActivity(guildID, game, checkAfter, duration)
	if err != nil {
		slog.Error("failed to get player activity",
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Failed to get player activity: " + err.Error()),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
	}
	slog.Debug("Player activity retrieved",
		slog.String("guild_id", i.GuildID),
		slog.String("game", game),
		slog.String("after", after),
		slog.Time("check_after", checkAfter),
		slog.Int("total_players", activity.ActivePlayers+activity.InactivePlayers),
		slog.String("active_players", p.Sprintf("%d (%.2f%%)", activity.ActivePlayers, activity.ActivePercentage)),
		slog.String("inactive_players", p.Sprintf("%d (%.2f%%)", activity.InactivePlayers, activity.InactivePercentage)),
	)

	embeds := []*discordgo.MessageEmbed{
		{
			Title:  "Player Activity for " + game,
			Fields: []*discordgo.MessageEmbedField{},
		},
	}
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "After",
		Value:  timeToString(after),
		Inline: false,
	})
	if since != "" {
		embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
			Name:   "Since",
			Value:  timeToString(since),
			Inline: false,
		})
	}
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "Total Players",
		Value:  p.Sprintf("%d", activity.ActivePlayers+activity.InactivePlayers),
		Inline: false,
	})
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "Active Players",
		Value:  p.Sprintf("%d (%.2f%%)", activity.ActivePlayers, activity.ActivePercentage),
		Inline: false,
	})
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "Inactive Players",
		Value:  p.Sprintf("%d (%.2f%%)", activity.InactivePlayers, activity.InactivePercentage),
		Inline: false,
	})

	resp := disgomsg.NewResponse(
		disgomsg.WithEmbeds(embeds),
	)

	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("failed to send response",
			slog.Any("error", err),
		)
	}
}

func playerChurn(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	var game, after, since string
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		switch option.Name {
		case "game":
			game = option.StringValue()
		case "after":
			after = option.StringValue()
		case "since":
			since = option.StringValue()
		}
	}

	slog.Debug("Player churn command received",
		slog.String("guild_id", i.GuildID),
		slog.String("game", game),
		slog.String("after", after),
		slog.String("since", since),
	)

	duration := getDuration(after)
	checkAfter := getTime(since)

	// guildID := i.GuildID
	guildID := "236523452230533121"
	churn, err := GetPlayerChurn(guildID, game, checkAfter, duration)
	if err != nil {
		slog.Error("failed to get player churn",
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Failed to get player churn: " + err.Error()),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
	}
	slog.Debug("Player churn retrieved",
		slog.String("guild_id", i.GuildID),
		slog.String("game", game),
		slog.String("after", after),
		slog.Time("check_after", checkAfter),
		slog.Int("total_players", churn.ActivePlayers+churn.InactivePlayers),
		slog.String("active_players", p.Sprintf("%d (%.2f%%)", churn.ActivePlayers, churn.ActivePercentage)),
		slog.String("inactive_players", p.Sprintf("%d (%.2f%%)", churn.InactivePlayers, churn.InactivePercentage)),
	)

	embeds := []*discordgo.MessageEmbed{
		{
			Title:  "Player Churn for " + game,
			Fields: []*discordgo.MessageEmbedField{},
		},
	}
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "After",
		Value:  timeToString(after),
		Inline: false,
	})
	if since != "" {
		embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
			Name:   "Since",
			Value:  timeToString(since),
			Inline: false,
		})
	}
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "Total Players",
		Value:  p.Sprintf("%d", churn.ActivePlayers+churn.InactivePlayers),
		Inline: false,
	})
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "Active Players",
		Value:  p.Sprintf("%d (%.2f%%)", churn.ActivePlayers, churn.ActivePercentage),
		Inline: false,
	})
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "Inactive Players",
		Value:  p.Sprintf("%d (%.2f%%)", churn.InactivePlayers, churn.InactivePercentage),
		Inline: false,
	})

	resp := disgomsg.NewResponse(
		disgomsg.WithEmbeds(embeds),
	)

	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("failed to send response",
			slog.Any("error", err),
		)
	}

}
