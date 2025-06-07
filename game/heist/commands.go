package heist

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/olekukonko/tablewriter"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/internal/channel"
	"github.com/rbrabson/goblin/internal/format"
	"github.com/rbrabson/goblin/internal/unicode"
)

const (
	MAX_WINNINGS_PER_PAGE = 30
)

// componentHandlers are the buttons that appear on messages sent by this bot.
var (
	componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"join_heist": joinHeist,
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"heist":       heist,
		"heist-admin": heistAdmin,
	}

	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "heist-admin",
			Description: "Heist admin commands.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "clear",
					Description: "Clears the criminal settings for the user.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "id",
							Description: "ID of the player to clear",
							Required:    true,
						},
					},
				},
				{
					Name:        "config",
					Description: "Configures the Heist bot.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "info",
							Description: "Returns the configuration information for the server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
						{
							Name:        "bail",
							Description: "Sets the base cost of bail.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "amount",
									Description: "The base cost of bail.",
									Required:    true,
								},
							},
						},
						{
							Name:        "cost",
							Description: "Sets the cost to plan or join a heist.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "amount",
									Description: "The cost to plan or join a heist.",
									Required:    true,
								},
							},
						},
						{
							Name:        "death",
							Description: "Sets how long players remain dead.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "time",
									Description: "The time the player remains dead, in seconds.",
									Required:    true,
								},
							},
						},
						{
							Name:        "patrol",
							Description: "Sets the time the authorities will prevent a new heist.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "time",
									Description: "The time the authorities will patrol, in seconds.",
									Required:    true,
								},
							},
						},
						{
							Name:        "sentence",
							Description: "Sets the base apprehension time when caught.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "time",
									Description: "The base time, in seconds.",
									Required:    true,
								},
							},
						},
						{
							Name:        "wait",
							Description: "Sets how long players can gather others for a heist.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "time",
									Description: "The time to wait for players to join the heist, in seconds.",
									Required:    true,
								},
							},
						},
					},
				},
				{
					Name:        "theme",
					Description: "Commands that interact with the heist themes.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "list",
							Description: "Gets the list of available heist themes.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
						{
							Name:        "set",
							Description: "Sets the current heist theme.",
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "name",
									Description: "Name of the theme to set.",
									Required:    true,
								},
							},
							Type: discordgo.ApplicationCommandOptionSubCommand,
						},
					},
				},
				{
					Name:        "reset",
					Description: "Resets a new heist that is hung.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "vault-reset",
					Description: "Resets the vaults to their maximum value.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}

	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "heist",
			Description: "Heist game commands.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "bail",
					Description: "Bail a player out of jail.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "id",
							Description: "ID of the player to bail. Defaults to you.",
							Required:    false,
						},
					},
				},
				{
					Name:        "stats",
					Description: "Shows a user's stats.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "start",
					Description: "Plans a new heist.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "targets",
					Description: "Gets the list of available heist targets.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}
)

// config routes the configuration commands to the proper handlers.
func config(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options[0].Options
	switch options[0].Name {
	case "cost":
		configCost(s, i)
	case "sentence":
		configSentence(s, i)
	case "patrol":
		configPatrol(s, i)
	case "bail":
		configBail(s, i)
	case "death":
		configDeath(s, i)
	case "wait":
		configWait(s, i)
	case "info":
		configInfo(s, i)
	}
}

// heistAdmin routes the commands to the subcommand and subcommandgroup handlers
func heistAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
	case "clear":
		clearMember(s, i)
	case "config":
		config(s, i)
	case "reset":
		resetHeist(s, i)
	case "vault-reset":
		resetVaults(s, i)
	case "theme":
		theme(s, i)
	}
}

// heist routes the commands to the subcommand and subcommandgroup handlers
func heist(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "bail":
		bailoutPlayer(s, i)
	case "start":
		planHeist(s, i)
	case "stats":
		playerStats(s, i)
	case "targets":
		listTargets(s, i)
	}
}

// theme routes the theme commands to the proper handlers.
func theme(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options[0].Options
	switch options[0].Name {
	case "list":
		listThemes(s, i)
	case "set":
		setTheme(s, i)
	}
}

// planHeist plans a new heist
func planHeist(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Create a new heist
	heist, err := NewHeist(i.GuildID, i.Member.User.ID)
	if err != nil {
		slog.Warn("unable to create the heist",
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
		return
	}

	theme := GetTheme(i.GuildID)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent("Planning a " + theme.Heist + " heist..."),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("failed to send response",
			slog.Any("error", err),
		)
	}
	heist.Organizer.guildMember.SetName(i.Member.User.Username, i.Member.Nick, i.Member.User.GlobalName)
	heist.interaction = i

	// The organizer has to pay a fee to plan the heist.
	account := bank.GetAccount(i.GuildID, i.Member.User.ID)
	if err := account.Withdraw(heist.config.HeistCost); err != nil {
		slog.Error("failed to withdraw heist cost",
			slog.Any("error", err),
		)
	}

	if err := heistMessage(s, heist); err != nil {
		slog.Error("failed to send heist message",
			slog.Any("error", err),
		)
	}

	waitForHeistToStart(s, i, heist)

	if len(heist.Crew) < 2 {
		heist.Cancel()

		if err := heistMessage(s, heist); err != nil {
			slog.Error("failed to send heist message",
				slog.Any("error", err),
			)
		}
		p := message.NewPrinter(language.AmericanEnglish)
		msg := disgomsg.NewMessage(
			disgomsg.WithContent(p.Sprintf("The %s was cancelled due to lack of interest.", heist.theme.Heist)),
		)
		if _, err := msg.Send(s, i.ChannelID); err != nil {
			slog.Error("failed to send message",
				slog.Any("error", err),
			)
		}
		slog.Info("Heist cancelled due to lack of interest",
			slog.String("guild", heist.GuildID),
			slog.String("heist", heist.theme.Heist),
		)

		return
	}

	res, err := heist.Start()
	if err != nil {
		slog.Error("unable to start the heist",
			slog.String("guildID", heist.GuildID),
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
		)
		if err := resp.Send(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
		return
	}
	defer heist.End()

	slog.Debug("heist is starting",
		slog.String("guildID", heist.GuildID),
	)

	mute := channel.NewChannelMute(s, i)
	mute.MuteChannel()
	defer mute.UnmuteChannel()

	err = heistMessage(s, heist)
	if err != nil {
		slog.Error("unable to mark the heist message as started",
			slog.String("guildID", heist.GuildID),
			slog.Any("error", err),
		)
	}

	p := message.NewPrinter(language.AmericanEnglish)
	msg := disgomsg.NewMessage(
		disgomsg.WithContent(p.Sprintf("The %s is starting with %d members.", heist.theme.Heist, len(heist.Crew))),
	)
	if _, err := msg.Send(s, i.ChannelID); err != nil {
		slog.Error("failed to send message",
			slog.String("channelID", i.ChannelID),
			slog.Any("error", err),
		)
	}

	time.Sleep(3 * time.Second)
	if err := heistMessage(s, heist); err != nil {
		slog.Error("failed to send heist message",
			slog.Any("error", err),
		)
	}

	sendHeistResults(s, i, res)

	res.Target.StealFromValut(res.TotalStolen)
}

// waitForHeistToStart waits until the planning stage for the heist expires.
func waitForHeistToStart(s *discordgo.Session, i *discordgo.InteractionCreate, heist *Heist) {
	// Wait for the heist to be ready to start
	waitTime := heist.StartTime.Add(heist.config.WaitTime)
	slog.Debug("wait for heist to start",
		slog.String("guildID", heist.GuildID),
		slog.Time("waitTime", waitTime),
		slog.Duration("configWaitTime", heist.config.WaitTime),
		slog.Time("currentTime", time.Now()),
	)
	for !time.Now().After(waitTime) {
		maximumWait := time.Until(waitTime)
		timeToWait := min(maximumWait, time.Duration(5*time.Second))
		if timeToWait < 0 {
			slog.Debug("wait for the heist to start is over",
				slog.String("guildID", heist.GuildID),
				slog.Duration("maximumWait", maximumWait),
				slog.Duration("timeToWait", timeToWait),
			)
			break
		}
		time.Sleep(timeToWait)

		if err := heistMessage(s, heist); err != nil {
			slog.Error("failed to send heist message",
				slog.Any("error", err),
			)
		}
	}
}

// sendHeistResults sends the results of the heist to the channel
func sendHeistResults(s *discordgo.Session, i *discordgo.InteractionCreate, res *HeistResult) {
	p := message.NewPrinter(language.AmericanEnglish)
	theme := GetTheme(i.GuildID)

	slog.Debug("hitting "+res.Target.Name,
		slog.String("guildID", i.GuildID),
	)
	msg := p.Sprintf("The %s has decided to hit **%s**.", theme.Crew, res.Target.Name)
	if _, err := s.ChannelMessageSend(i.ChannelID, msg); err != nil {
		slog.Error("failed to send message",
			slog.String("channelID", i.ChannelID),
			slog.Any("error", err),
		)
	}
	time.Sleep(3 * time.Second)

	// Process the results
	for _, result := range res.AllResults {
		guildMember := result.Player.guildMember
		msg = p.Sprintf(result.Message+"\n", "**"+guildMember.Name+"**")
		if result.Status == APPREHENDED {
			msg += p.Sprintf("`%s dropped out of the game.`", guildMember.Name)
		}
		if _, err := s.ChannelMessageSend(i.ChannelID, msg); err != nil {
			slog.Error("failed to send message",
				slog.String("channelID", i.ChannelID),
				slog.Any("error", err),
			)
		}
		time.Sleep(3 * time.Second)
	}

	if len(res.Escaped) == 0 {
		msg = "\nNo one made it out safe."
		if _, err := s.ChannelMessageSend(i.ChannelID, msg); err != nil {
			slog.Error("failed to send message",
				slog.String("channelID", i.ChannelID),
				slog.Any("error", err),
			)
		}
	} else {
		msg = "\nThe raid is now over. Distributing player spoils."
		if _, err := s.ChannelMessageSend(i.ChannelID, msg); err != nil {
			slog.Error("failed to send message",
				slog.String("channelID", i.ChannelID),
				slog.Any("error", err),
			)
		}

		// Render the results into a table and returnt he results.
		var tableBuffer strings.Builder
		table := tablewriter.NewWriter(&tableBuffer)
		table.SetBorder(false)
		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(true)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetHeaderLine(false)
		table.SetBorder(false)
		table.SetTablePadding("\t")
		table.SetNoWhiteSpace(true)
		table.SetHeader([]string{"Player", "Loot", "Bonus", "Total"})
		for _, result := range res.AllResults {
			guildMember := result.Player.guildMember
			if result.Status == FREE || result.Status == APPREHENDED {
				data := []string{guildMember.Name, p.Sprintf("%d", result.StolenCredits), p.Sprintf("%d", result.BonusCredits), p.Sprintf("%d", result.StolenCredits+result.BonusCredits)}
				table.Append(data)
				if table.NumLines() >= MAX_WINNINGS_PER_PAGE {
					table.Render()
					if _, err := s.ChannelMessageSend(i.ChannelID, "```\n"+tableBuffer.String()+"\n```"); err != nil {
						slog.Error("failed to send message",
							slog.String("channelID", i.ChannelID),
							slog.Any("error", err),
						)
					}
					table.ClearRows()
					tableBuffer.Reset()
				}
			}
		}
		if table.NumLines() > 0 {
			table.Render()
			if _, err := s.ChannelMessageSend(i.ChannelID, "```\n"+tableBuffer.String()+"```"); err != nil {
				slog.Error("failed to send message",
					slog.String("channelID", i.ChannelID),
					slog.Any("error", err),
				)
			}
		}
	}

	// Update the status for each player and then save the information
	for _, result := range res.AllResults {
		result.Player.heist = result.heist
		switch result.Status {
		case APPREHENDED:
			result.Player.Apprehended()
		case DEAD:
			result.Player.Died()
		default:
			result.Player.Escaped()
		}

		if len(res.Escaped) > 0 && result.StolenCredits != 0 {
			account := bank.GetAccount(i.GuildID, result.Player.MemberID)
			if err := account.Deposit(result.StolenCredits + result.BonusCredits); err != nil {
				slog.Error("failed to deposit stolen credits",
					slog.String("guildID", i.GuildID),
					slog.Any("error", err),
				)
			}
			slog.Info("heist Loot",
				slog.String("guildID", i.GuildID),
				slog.String("memberID", account.MemberID),
				slog.Int("stolen", result.StolenCredits),
				slog.Int("bonus", result.BonusCredits),
			)
		}
	}

	h := GetHeist(i.GuildID)
	h.End()

	if err := heistMessage(s, h); err != nil {
		slog.Error("failed to heist send message",
			slog.Any("error", err),
		)
	}
}

// joinHeist attempts to join a heist that is being planned
func joinHeist(s *discordgo.Session, i *discordgo.InteractionCreate) {
	heist := GetHeist(i.GuildID)
	if heist == nil {
		theme := GetTheme(i.GuildID)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("No %s is being planned", theme.Heist)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
		return
	}

	heistMember := getHeistMember(i.GuildID, i.Member.User.ID)
	heistMember.guildMember.SetName(i.Member.User.Username, i.Member.Nick, i.Member.User.GlobalName)
	err := heist.AddCrewMember(heistMember)
	if err != nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
		return
	}
	if err := heistMessage(s, heist); err != nil {
		slog.Error("failed to heist message",
			slog.Any("error", err),
		)
	}

	// Withdraw the cost of the heist from the player's account. We know the player already
	// has the required number of credits as this is verified when adding them to the heist.
	account := bank.GetAccount(i.GuildID, heistMember.MemberID)
	if err := account.Withdraw(heist.config.HeistCost); err != nil {
		slog.Error("failed to withdraw heist",
			slog.String("guildID", i.GuildID),
			slog.Any("error", err),
		)
	}

	heistMember.guildMember.SetName(i.Member.User.Username, i.Member.Nick, i.Member.User.GlobalName)
	p := message.NewPrinter(language.AmericanEnglish)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("You have joined the %s at a cost of %d credits.", heist.theme.Heist, heist.config.HeistCost)),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("failed to send response",
			slog.Any("error", err),
		)
	}
}

// playerStats shows a player's heist stats
func playerStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	theme := GetTheme(i.GuildID)
	player := getHeistMember(i.GuildID, i.Member.User.ID)
	player.guildMember.SetName(i.Member.User.Username, i.Member.Nick, i.Member.User.GlobalName)
	caser := cases.Caser(cases.Title(language.Und, cases.NoLower))

	account := bank.GetAccount(i.GuildID, i.Member.User.ID)

	var sentence string
	if player.Status == APPREHENDED {
		if player.RemainingJailTime() <= 0 {
			sentence = "Served"
		} else {
			timeRemaining := time.Until(player.JailTimer)
			sentence = format.Duration(timeRemaining)
		}
	} else {
		sentence = "None"
	}

	embeds := []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Title:       player.guildMember.Name,
			Description: player.CriminalLevel.String(),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Status",
					Value:  player.Status.String(),
					Inline: true,
				},
				{
					Name:   "Spree",
					Value:  p.Sprintf("%d", player.Spree),
					Inline: true,
				},
				{
					Name:   caser.String(theme.Bail),
					Value:  p.Sprintf("%d", player.BailCost),
					Inline: true,
				},
				{
					Name:   caser.String(theme.Sentence),
					Value:  sentence,
					Inline: true,
				},
				{
					Name:   "Apprehended",
					Value:  p.Sprintf("%d", player.JailCounter),
					Inline: true,
				},
				{
					Name:   "Total Deaths",
					Value:  p.Sprintf("%d", player.Deaths),
					Inline: true,
				},
				{
					Name:   "Lifetime Apprehensions",
					Value:  p.Sprintf("%d", player.TotalJail),
					Inline: true,
				},
				{
					Name:   "Credits",
					Value:  p.Sprintf("%d", account.CurrentBalance),
					Inline: true,
				},
			},
		},
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		slog.Error("unable to send the player stats to Discord",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
	}
}

// bailoutPlayer bails a player player out from jail. This defaults to the player initiating the command, but can
// be another player as well.
func bailoutPlayer(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var playerID string
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		if option.Name == "id" {
			playerID = strings.TrimSpace(option.StringValue())
		}
	}

	initiatingHeistMember := getHeistMember(i.GuildID, i.Member.User.ID)
	initiatingHeistMember.guildMember.SetName(i.Member.User.Username, i.Member.Nick, i.Member.User.GlobalName)
	account := bank.GetAccount(i.GuildID, i.Member.User.ID)

	var resp disgomsg.Response
	var heistMember *HeistMember
	if playerID != "" {
		heistMember = getHeistMember(i.GuildID, playerID)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Bailing out %s...", heistMember.guildMember.Name)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send response",
				slog.String("guildID", i.GuildID),
				slog.Any("error", err),
			)
		}
	} else {
		heistMember = initiatingHeistMember
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Bailing yourself out..."),
		)
		err := resp.SendEphemeral(s, i.Interaction)
		if err != nil {
			slog.Error("unable to send the bail message",
				slog.String("guildID", i.GuildID),
				slog.String("memberID", i.Member.User.ID),
				slog.Any("error", err),
			)
			return
		}
	}

	if heistMember.Status != APPREHENDED {
		var msg string
		if heistMember.MemberID == i.Member.User.ID {
			msg = "You are not in jail"
		} else {
			msg = fmt.Sprintf("%s is not in jail", heistMember.guildMember.Name)
		}
		if err := resp.WithContent(msg).Edit(s); err != nil {
			slog.Error("failed to edit the message",
				slog.String("guildID", i.GuildID),
				slog.Any("error", err),
			)
		}
		return
	}

	if heistMember.RemainingJailTime() <= 0 {
		if heistMember.MemberID == i.Member.User.ID {
			if err := resp.WithContent("You have already served your sentence.").Edit(s); err != nil {
				slog.Error("failed to send response",
					slog.String("guildID", i.GuildID),
					slog.String("memberID", i.Member.User.ID),
					slog.Any("error", err),
				)
			}
		} else {
			if err := resp.WithContent(fmt.Sprintf("%s has already served their sentence.", heistMember.guildMember.Name)).Edit(s); err != nil {
				slog.Error("failed to send response",
					slog.String("guildID", i.GuildID),
					slog.String("memberID", i.Member.User.ID),
					slog.Any("error", err),
				)
			}
		}
		heistMember.ClearJailAndDeathStatus()
		return
	}

	err := account.Withdraw(heistMember.BailCost)
	if err != nil {
		p := message.NewPrinter(language.AmericanEnglish)
		if err := resp.WithContent(p.Sprintf("You do not have enough credits to play the bail of %d", heistMember.BailCost)).Edit(s); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
		return
	}
	heistMember.Status = OOB

	if heistMember.MemberID == initiatingHeistMember.MemberID {
		p := message.NewPrinter(language.AmericanEnglish)
		if err := resp.WithContent(p.Sprintf("Congratulations, you are now free! You spent %d credits on your bail. Enjoy your freedom while it lasts.", heistMember.BailCost)).Edit(s); err != nil {
			slog.Error("failed to send response",
				slog.Any("error", err),
			)
		}
		return
	}

	member := guild.GetMember(heistMember.GuildID, heistMember.MemberID)
	initiatingMember := initiatingHeistMember.guildMember
	p := message.NewPrinter(language.AmericanEnglish)
	if err := resp.WithContent(p.Sprintf("Congratulations, %s, %s bailed you out by spending %d credits and now you are free!. Enjoy your freedom while it lasts.", member.Name, initiatingMember.Name, heistMember.BailCost)).Edit(s); err != nil {
		slog.Error("failed to send response",
			slog.Any("error", err),
		)
	}
}

// heistMessage sends the main command used to plan, join and leave a heist. It also handles the case where
// the heist starts, disabling the buttons to join/leave/cancel the heist.
func heistMessage(s *discordgo.Session, heist *Heist) error {
	heist.mutex.Lock()
	defer heist.mutex.Unlock()

	var status string
	var buttonDisabled bool
	switch heist.State {
	case Planning:
		until := time.Until(heist.StartTime.Add(heist.config.WaitTime))
		status = "Starts in " + format.Duration(until)
		buttonDisabled = false
	case InProgress:
		status = "Started"
		buttonDisabled = true
	case Cancelled:
		status = "Canceled"
		buttonDisabled = true
	case Completed:
		status = "Ended"
		buttonDisabled = true
	default:
		status = "Ended"
		buttonDisabled = true
	}

	crew := make([]string, 0, len(heist.Crew))
	for _, crewMember := range heist.Crew {
		crew = append(crew, crewMember.guildMember.Name)
	}

	caser := cases.Caser(cases.Title(language.Und, cases.NoLower))
	p := message.NewPrinter(language.AmericanEnglish)
	description := p.Sprintf("A new %s is being planned by %s. You can join the %s for a cost of %d credits at any time prior to the %s starting.",
		heist.theme.Heist,
		heist.Organizer.guildMember.Name,
		heist.theme.Heist,
		heist.config.HeistCost,
		heist.theme.Heist,
	)

	embeds := []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Title:       "Heist",
			Description: description,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Status",
					Value:  status,
					Inline: true,
				},
				{
					Name:   fmt.Sprintf("%s (%d members)", caser.String(heist.theme.Crew), len(crew)),
					Value:  strings.Join(crew, ", "),
					Inline: true,
				},
			},
		},
	}
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Join",
				Style:    discordgo.SuccessButton,
				Disabled: buttonDisabled,
				CustomID: "join_heist",
				Emoji:    nil,
			},
		}},
	}
	emptymsg := ""
	_, err := s.InteractionResponseEdit(heist.interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &embeds,
		Components: &components,
		Content:    &emptymsg,
	})
	if err != nil {
		slog.Error("unable to send the heist message",
			slog.String("guildID", heist.GuildID),
			slog.Any("error", err),
		)
		return err
	}

	return nil
}

/******** ADMIN COMMANDS ********/

// Reset resets the heist in case it hangs
func resetHeist(s *discordgo.Session, i *discordgo.InteractionCreate) {
	mute := channel.NewChannelMute(s, i)
	defer mute.UnmuteChannel()

	heistLock.Lock()
	heist := currentHeists[i.GuildID]
	delete(currentHeists, i.GuildID)
	heistLock.Unlock()

	if heist == nil {
		theme := GetTheme(i.GuildID)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("No %s is being planned", theme.Heist)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send the response",
				slog.Any("error", err),
			)
		}
		return
	}

	heist.Cancel()
	if err := heistMessage(s, heist); err != nil {
		slog.Error("unable to send heist message",
			slog.Any("error", err),
		)
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("The %s has been reset", heist.theme.Heist)),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("unable to send the response",
			slog.Any("error", err),
		)
	}
}

// resetVaults sets the vaults within the guild to their maximum value.
func resetVaults(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ResetVaultsToMaximumValue(i.GuildID)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent("Vaults have been reset to their maximum value"),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("unable to send the response",
			slog.Any("error", err),
		)
	}
}

// listTargets displays a list of available heist targets.
func listTargets(s *discordgo.Session, i *discordgo.InteractionCreate) {
	theme := GetTheme(i.GuildID)
	targets := GetTargets(i.GuildID, theme.Name)

	if len(targets) == 0 {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("There aren't any targets!"),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send the response",
				slog.Any("error", err),
			)
		}
		return
	}

	// Lets return the data in an Ascii table. Ideally, it would be using a Discord embed, but unfortunately
	// Discord only puts three columns per row, which isn't enough for our purposes.
	var tableBuffer strings.Builder
	table := tablewriter.NewWriter(&tableBuffer)
	table.SetBorder(false)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.SetHeader([]string{"ID", "Max Crew", theme.Vault, "Max " + theme.Vault, "Success Rate"})
	for _, target := range targets {
		data := []string{target.Name, fmt.Sprintf("%d", target.CrewSize), fmt.Sprintf("%d", target.Vault), fmt.Sprintf("%d", target.VaultMax), fmt.Sprintf("%.2f", target.Success)}
		table.Append(data)
	}
	table.Render()

	resp := disgomsg.NewResponse(
		disgomsg.WithContent("```\n" + tableBuffer.String() + "\n```"),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("failed to send the response",
			slog.Any("error", err),
		)
	}
}

// clearMember clears the criminal state of the player.
func clearMember(s *discordgo.Session, i *discordgo.InteractionCreate) {
	memberID := i.ApplicationCommandData().Options[0].Options[0].StringValue()
	member := guild.GetMember(i.GuildID, memberID)
	heistMember := getHeistMember(i.GuildID, memberID)
	heistMember.ClearJailAndDeathStatus()
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Cleared %s's criminal record", member.Name)),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("failed to send the response",
			slog.Any("error", err),
		)
	}
}

// listThemes returns the list of available themes that may be used for heists
func listThemes(s *discordgo.Session, i *discordgo.InteractionCreate) {
	themes, err := GetThemeNames(i.GuildID)
	if err != nil {
		slog.Warn("Unable to get the themes",
			slog.String("guildID", i.GuildID),
			slog.Any("error", err),
		)
	}

	embeds := []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Title:       "Available Themes",
			Description: "Available Themes for the Heist bot",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Themes",
					Value:  strings.Join(themes[:], ", "),
					Inline: true,
				},
			},
		},
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		slog.Error("Unable to send list of themes to the user for `list themes`",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
	}
}

// setTheme sets the heist theme to the one specified in the command
func setTheme(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var themeName string
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	for _, option := range options {
		if option.Name == "name" {
			themeName = strings.TrimSpace(option.StringValue())
		}
	}

	config := GetConfig(i.GuildID)
	if themeName == config.Theme {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Theme `" + themeName + "` is already being used."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send the response",
				slog.Any("error", err),
			)
		}
		return
	}
	theme := GetTheme(i.GuildID)
	config.Theme = theme.Name
	slog.Debug("now using theme ",
		slog.String("theme", config.Theme),
	)

	resp := disgomsg.NewResponse(
		disgomsg.WithContent("Theme `" + themeName + "` is now being used."),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("failed to send the response",
			slog.Any("error", err),
		)
	}
}

// configCost sets the cost to plan or join a heist
func configCost(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := GetConfig(i.GuildID)
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	cost := options[0].IntValue()
	config.HeistCost = int(cost)

	p := message.NewPrinter(language.AmericanEnglish)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Cost set to %d", cost)),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("failed to send the response",
			slog.Any("error", err),
		)
	}
	writeConfig(config)
}

// configSentence sets the base aprehension time when a player is apprehended.
func configSentence(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options[0]
	if options == nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("No sentence time provided (option missing)"),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send the response",
				slog.Any("error", err),
			)
		}
		return
	}
	options = options.Options[0]
	if options == nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("No sentence time provided (1st level option missing)"),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send the response",
				slog.Any("error", err),
			)
		}
		return
	}

	config := GetConfig(i.GuildID)
	if i.ApplicationCommandData().Options[0].Options[0].IntValue() == 0 {
		config.SentenceBase = 0
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Sentence disabled"),
		)
		if err := resp.Send(s, i.Interaction); err != nil {
			slog.Error("failed to send the response",
				slog.Any("error", err),
			)
		}
		writeConfig(config)
		return
	}
	sentence := i.ApplicationCommandData().Options[0].Options[0].IntValue()
	config.SentenceBase = time.Duration(sentence * int64(time.Second))

	p := message.NewPrinter(language.AmericanEnglish)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Sentence set to %d seconds", sentence)),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("unable to send the response",
			slog.Any("error", err),
		)
	}

	writeConfig(config)
}

// configPatrol sets the time authorities will prevent a new heist following one being completed.
func configPatrol(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := GetConfig(i.GuildID)
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	patrol := options[0].IntValue()
	config.PoliceAlert = time.Duration(patrol * int64(time.Second))

	p := message.NewPrinter(language.AmericanEnglish)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Patrol set to %d", patrol)),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("failed to send the response",
			slog.Any("error", err),
		)
	}

	writeConfig(config)
}

// configBail sets the base cost of bail.
func configBail(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := GetConfig(i.GuildID)
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	bail := options[0].IntValue()
	config.BailBase = int(bail)

	p := message.NewPrinter(language.AmericanEnglish)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Bail set to %d", bail)),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("failed to send the response",
			slog.Any("error", err),
		)
	}

	writeConfig(config)
}

// configDeath sets how long players remain dead.
func configDeath(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := GetConfig(i.GuildID)
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	death := options[0].IntValue()
	config.PoliceAlert = time.Duration(death * int64(time.Second))

	p := message.NewPrinter(language.AmericanEnglish)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Death set to %d", death)),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("failed to send the response",
			slog.Any("error", err),
		)
	}

	writeConfig(config)
}

// configWait sets how long players wait for others to join the heist.
func configWait(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := GetConfig(i.GuildID)
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	wait := options[0].IntValue()
	config.WaitTime = time.Duration(wait * int64(time.Second))

	p := message.NewPrinter(language.AmericanEnglish)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Wait set to %d", wait)),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("failed to send the resposne",
			slog.Any("error", err),
		)
	}

	writeConfig(config)
}

// configInfo returns the configuration for the Heist bot on this server.
func configInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := GetConfig(i.GuildID)

	embed := &discordgo.MessageEmbed{
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "bail",
				Value:  fmt.Sprintf("%d", config.BailBase),
				Inline: true,
			},
			{
				Name:   "cost",
				Value:  fmt.Sprintf("%d", config.HeistCost),
				Inline: true,
			},
			{
				Name:   "death",
				Value:  fmt.Sprintf("%.f", config.DeathTimer.Seconds()),
				Inline: true,
			},
			{
				Name:   "patrol",
				Value:  fmt.Sprintf("%.f", config.PoliceAlert.Seconds()),
				Inline: true,
			},
			{
				Name:   "sentence",
				Value:  fmt.Sprintf("%.f", config.SentenceBase.Seconds()),
				Inline: true,
			},
			{
				Name:   "wait",
				Value:  fmt.Sprintf("%.f", config.WaitTime.Seconds()),
				Inline: true,
			},
		},
	}

	embeds := []*discordgo.MessageEmbed{
		embed,
	}
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Heist Configuration",
			Embeds:  embeds,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		slog.Error("unable to send a response for `config info`",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
	}
}
