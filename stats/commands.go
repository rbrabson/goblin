package stats

import (
	"log/slog"
	"math"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	All   = "all"
	Heist = "heist"
	Race  = "race"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"stats-admin": statsAdmin,
		"stats":       stats,
	}

	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "stats-admin",
			Description: "Commands used to interact with the stats system.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "retention",
					Description: "View player retention.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "game",
							Description: "The game for which to determine the retention.",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "All",
									Value: All,
								},
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
							Description: "The time period to get the number of games.",
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
							Description: "The time period to check the player retention.",
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
					Name:        "played",
					Description: "View the number of games played.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "game",
							Description: "The game for which to get the number of games played.",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "All",
									Value: All,
								},
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
							Name:        "since",
							Description: "The time period to check the number of games played.",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    false,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Yesterday",
									Value: OneDayAgo,
								},
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

	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "stats",
			Description: "Commands used to interact with the stats system.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "played",
					Description: "View games played by a player.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "game",
							Description: "The game for which to determine the number of games played.",
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

	// TODO: playerActivity isn't being used, so it can be removed.
	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "retention":
		playerRetention(s, i)
	case "played":
		gamesPlayed(s, i)
	}
}

// playerRetention handles the /stats-admin retention command.
func playerRetention(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)
	titleCaser := cases.Title(language.AmericanEnglish)

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

	slog.Debug("player retention command received",
		slog.String("guild_id", i.GuildID),
		slog.String("game", game),
		slog.String("after", after),
		slog.String("since", since),
	)

	guildID := getGuildID(i)

	firstGameDate := getFirstGameDate(guildID, game)
	duration := getDuration(guildID, game, after, firstGameDate)
	cuttoff := getTime(guildID, game, since, firstGameDate)

	slog.Debug("duration and cutoff calculated",
		slog.String("guild_id", i.GuildID),
		slog.String("game", game),
		slog.Time("cutoff", cuttoff),
	)

	var retention *PlayerRetention
	var err error
	if game == "" || game == "all" {
		retention, err = GetPlayerRetention(guildID, cuttoff, duration)
	} else {
		retention, err = GetPlayerRetentionForGame(guildID, game, cuttoff, duration)
	}
	if err != nil {
		slog.Error("failed to get player retention",
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Failed to get player retention: " + err.Error()),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
	}
	slog.Debug("player retention retrieved",
		slog.String("guild_id", i.GuildID),
		slog.String("game", game),
		slog.String("after", after),
		slog.Time("cutoff", cuttoff),
		slog.Int("total_players", retention.ActivePlayers+retention.InactivePlayers),
		slog.String("active_players", p.Sprintf("%d (%.2f%%)", retention.ActivePlayers, retention.ActivePercentage)),
		slog.String("inactive_players", p.Sprintf("%d (%.2f%%)", retention.InactivePlayers, retention.InactivePercentage)),
	)

	var embeds []*discordgo.MessageEmbed

	if game == "" || game == "all" {
		embeds = []*discordgo.MessageEmbed{
			{
				Title:  titleCaser.String("Player Retention"),
				Fields: []*discordgo.MessageEmbedField{},
			},
		}
	} else {
		embeds = []*discordgo.MessageEmbed{
			{
				Title:  titleCaser.String("Player Retention for " + game),
				Fields: []*discordgo.MessageEmbedField{},
			},
		}
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
		Value:  p.Sprintf("%d", retention.ActivePlayers+retention.InactivePlayers),
		Inline: false,
	})
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "Active Players",
		Value:  p.Sprintf("%d (%.0f%%)", retention.ActivePlayers, retention.ActivePercentage),
		Inline: false,
	})
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "Inactive Players",
		Value:  p.Sprintf("%d (%.0f%%)", retention.InactivePlayers, retention.InactivePercentage),
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

// gamesPlayed handles the /stats-admin games command.
func gamesPlayed(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)
	titleCaser := cases.Title(language.AmericanEnglish)
	today := today()

	var game, since string
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		switch option.Name {
		case "game":
			game = option.StringValue()
		case "since":
			since = option.StringValue()
		}
	}

	slog.Debug("games played command received",
		slog.String("guild_id", i.GuildID),
		slog.String("game", game),
		slog.String("since", since),
	)

	guildID := getGuildID(i)

	firstGameDate := getFirstServerGameDate(guildID, game)
	checkAfter := getTime(guildID, game, since, firstGameDate)

	var gamesPlayed *GamesPlayed
	var err error
	if game == "" {
		game = "all"
	}
	gamesPlayed, err = GetGamesPlayed(guildID, game, checkAfter, today)

	if err != nil {
		slog.Error("failed to get games played",
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Failed to get games played: " + err.Error()),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
	}
	slog.Debug("games played retrieved",
		slog.String("guild_id", i.GuildID),
		slog.String("game", game),
		slog.Time("check_after", checkAfter),
		slog.String("games", p.Sprintf("%d (%.2f%%)", gamesPlayed.TotalGamesPlayed, gamesPlayed.AverageGamesPerDay)),
		slog.String("unique_players", p.Sprintf("%d (%.2f%%)", gamesPlayed.UniquePlayers, gamesPlayed.AverageGamesPerPlayerPerDay)),
		slog.String("avg_games_per_player", p.Sprintf("%.2f", gamesPlayed.AverageGamesPerPlayer)),
	)

	var embeds []*discordgo.MessageEmbed
	if game == "" || game == "all" {
		embeds = []*discordgo.MessageEmbed{
			{
				Title:  titleCaser.String("Games Played"),
				Fields: []*discordgo.MessageEmbedField{},
			},
		}
	} else {
		embeds = []*discordgo.MessageEmbed{
			{
				Title:  titleCaser.String("Games Played for " + game),
				Fields: []*discordgo.MessageEmbedField{},
			},
		}
	}
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "Since",
		Value:  p.Sprintf("%s Ago", fmtDuration(today.Sub(checkAfter))),
		Inline: false,
	})
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "Total Games",
		Value:  p.Sprintf("%d", gamesPlayed.TotalGamesPlayed),
		Inline: false,
	})
	if checkAfter.Before(today) {
		embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
			Name:   "Average Games Per Day",
			Value:  p.Sprintf("%.0f", math.Round(gamesPlayed.AverageGamesPerDay)),
			Inline: false,
		})
	}
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "Total Players",
		Value:  p.Sprintf("%d", gamesPlayed.TotalPlayers),
		Inline: false,
	})
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "Average Players Per Day",
		Value:  p.Sprintf("%.0f", math.Round(gamesPlayed.TotalPlayersPerDay)),
		Inline: false,
	})
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "Unique Players For Every Day",
		Value:  p.Sprintf("%d", gamesPlayed.UniquePlayers),
		Inline: false,
	})
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "Average Unique Players Per Day",
		Value:  p.Sprintf("%.0f", math.Round(gamesPlayed.UniquePlayersPerDay)),
		Inline: false,
	})
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "Average Players Per Game",
		Value:  p.Sprintf("%.2f", gamesPlayed.AveragePlayersPerGame),
		Inline: false,
	})
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "Average Games Per Player",
		Value:  p.Sprintf("%.2f", gamesPlayed.AverageGamesPerPlayer),
		Inline: false,
	})
	embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
		Name:   "Average Games Per Player Per Day",
		Value:  p.Sprintf("%.2f", gamesPlayed.AverageGamesPerPlayerPerDay),
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

// stats handles the /stats command.
func stats(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	// TODO: playerActivity isn't being used, so it can be removed.
	options := i.ApplicationCommandData().Options
	if options[0].Name == "played" {
		playerGames(s, i)
	}
}

// playerGames handles the /stats played command.
func playerGames(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)
	titleCaser := cases.Title(language.AmericanEnglish)
	today := today()

	memberID := i.Member.User.ID
	var game string
	var member *guild.Member
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		if option.Name == "user" {
			var err error
			member, err = guild.GetMemberByUser(s, i.GuildID, option.UserValue(s))
			if err != nil {
				resp := disgomsg.NewResponse(
					disgomsg.WithContent("The user to get the account for was not found. Please try again."),
				)
				if err := resp.SendEphemeral(s, i.Interaction); err != nil {
					slog.Error("error sending response",
						slog.String("guildID", i.GuildID),
						slog.String("error", err.Error()),
					)
				}
				return
			}
			memberID = member.MemberID
		}
		if option.Name == "game" {
			game = option.StringValue()
		}
	}

	guildID := getGuildID(i)
	var guildMember *guild.Member
	if member == nil {
		guildMember = guild.GetMember(guildID, memberID).SetName(i.Member.User.Username, i.Member.Nick, i.Member.User.GlobalName)
	} else {
		guildMember = guild.GetMember(guildID, memberID).SetName(member.UserName, member.NickName, member.GlobalName)
	}

	ps, _ := readPlayerStats(guildID, memberID, game)
	if ps == nil {
		content := p.Sprintf("No player stats found for %s in the %s game.", guildMember.Name, game)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(content),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
		return
	}

	embeds := []*discordgo.MessageEmbed{
		{
			Title: titleCaser.String("Games Played for " + game),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Member",
					Value:  p.Sprintf("%s", guildMember.Name),
					Inline: false,
				},
				{
					Name:   "First Played",
					Value:  p.Sprintf("%s Ago", fmtDuration(today.Sub(ps.FirstPlayed))),
					Inline: false,
				},
				{
					Name:   "Last Played",
					Value:  p.Sprintf("%s Ago", fmtDuration(today.Sub(ps.LastPlayed))),
					Inline: false,
				},
				{
					Name:   "Games Played",
					Value:  p.Sprintf("%d", ps.NumberOfTimesPlayed),
					Inline: false,
				},
			},
		},
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithEmbeds(embeds),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("failed to send response",
			slog.Any("error", err),
		)
	}
}

// getGuildID returns the guild ID from the interaction.
func getGuildID(i *discordgo.InteractionCreate) string {
	guildID := i.GuildID
	// guildID := "236523452230533121"
	return guildID
}
