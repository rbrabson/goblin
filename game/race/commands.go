package race

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/internal/format"
	"github.com/rbrabson/goblin/internal/unicode"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	betButtons     = make(map[string]map[string]*raceButton) // guild -> label -> button
	betButtonMutex = sync.Mutex{}

	componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"join_race": joinRace,
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"race":       race,
		"race-admin": raceAdmin,
	}

	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "race",
			Description: "Race game commands.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "start",
					Description: "Starts a new race.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "stats",
					Description: "Returns the race stats for the player.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}

	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "race-admin",
			Description: "Race game admin commands.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "reset",
					Description: "Resets a hung race.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}
)

type raceButton struct {
	label string
	racer *RaceParticipant
}

// raceAdmin routes various `race-raceAdmin` subcommands to the appropriate handlers.
func raceAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		disgomsg.NewResponse(disgomsg.WithContent("System is shutting down")).SendEphemeral(s, i.Interaction)
		return
	}

	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		disgomsg.NewResponse(disgomsg.WithContent("You do not have permission to use this command.")).SendEphemeral(s, i.Interaction)
		return
	}

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "reset":
		resetRace(s, i)
	default:
		slog.Error("unknown command", slog.String("guildID", i.GuildID), slog.String("memberID", i.Member.User.ID), slog.String("command", options[0].Name))
		disgomsg.NewResponse(disgomsg.WithContent("Command is unknown")).SendEphemeral(s, i.Interaction)
	}
}

// race routes the various `race` subcommands to the appropriate handlers.
func race(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		disgomsg.NewResponse(disgomsg.WithContent("System is shutting down")).SendEphemeral(s, i.Interaction)
		return
	}

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "start":
		startRace(s, i)
	case "stats":
		raceStats(s, i)
	default:
		disgomsg.NewResponse(disgomsg.WithContent("Command is unknown")).SendEphemeral(s, i.Interaction)
		slog.Error("unknown command", slog.String("guildID", i.GuildID), slog.String("memberID", i.Member.User.ID), slog.String("command", options[0].Name))
	}
}

// resetRace resets a hung race.
func resetRace(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ResetRace(i.GuildID)
	disgomsg.NewResponse(disgomsg.WithContent("Race has been reset")).SendEphemeral(s, i.Interaction)
}

// startRace starts a race that other members may join.
func startRace(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// No need to worry about locking the race until the first raceMessage is sent, as no other members
	// can join until the message with the JOIN button is sent
	race, err := CreateNewRace(i.GuildID)
	if err != nil {
		slog.Warn("failed to create new race", slog.String("guildID", i.GuildID), slog.String("memberID", i.Member.User.ID), slog.Any("error", err))
		disgomsg.NewResponse(disgomsg.WithContent(unicode.FirstToLower(err.Error()))).SendEphemeral(s, i.Interaction)
		return
	}
	defer race.End()
	slog.Info("race created", slog.String("guiildID", i.GuildID), slog.String("memberID", i.Member.User.ID))

	race.interaction = i
	guildMember := guild.GetMember(i.GuildID, i.Member.User.ID).SetName(i.Member.User.Username, i.Member.Nick, i.Member.User.GlobalName)
	racer := getRaceMember(i.GuildID, guildMember)
	race.addRaceParticipant(racer)

	raceMessage(s, race, "start")

	slog.Info("waiting for members to join the race", slog.String("guildID", i.GuildID))
	waitForMembersToJoin(s, race)

	if len(race.Racers) < race.config.MinNumRacers {
		slog.Info("race cancelled due to insufficient racers", slog.String("guildID", i.GuildID), slog.Int("racers", len(race.Racers)))
		raceMessage(s, race, "cancelled")
		return
	}

	race.setState(RaceWaitingForBets)
	raceMessage(s, race, "betting")
	defer removeBetButtons(race)
	slog.Info("waiting for bets", slog.String("guildID", i.GuildID), slog.Int("racers", len(race.Racers)))
	waitForBetsToBePlaced(s, race)

	race.setState(RaceInProgress)
	raceMessage(s, race, "started")
	slog.Info("race starting", slog.String("guildID", i.GuildID), slog.Int("racers", len(race.Racers)), slog.Int("betsPlaced", len(race.Betters)))
	race.runRace(len([]rune(race.config.Track)))
	sendRaceLegs(s, race)

	race.setState(RaceFinished)
	raceMessage(s, race, "ended")
	slog.Info("race ended", slog.String("guildID", i.GuildID))

	sendRaceResults(s, i.ChannelID, race)
}

// waitForMembersToJoin waits until members join the race before proceeding
// to taking bets
func waitForMembersToJoin(s *discordgo.Session, race *Race) {
	memberJoinTime := time.Now().Add(race.config.WaitToStart)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	count := 0
	raceMessage(s, race, "update")
	for range ticker.C {
		count++
		if count%5 == 0 {
			raceMessage(s, race, "update")
		}
		if time.Until(memberJoinTime) <= 0 {
			break
		}
		if len(race.Racers) >= race.config.MaxNumRacers {
			break
		}
	}
}

// waitForBetsToBePlaced waits until bets are placed before starting the race.
func waitForBetsToBePlaced(s *discordgo.Session, race *Race) {
	betEndTime := time.Now().Add(race.config.WaitForBets)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	count := 0
	raceMessage(s, race, "betting")
	for range ticker.C {
		count++
		if count%5 == 0 {
			raceMessage(s, race, "betting")
		}
		if time.Until(betEndTime) <= 0 {
			break
		}
	}
}

// joinRace attempts to join a race that is getting ready to start.
func joinRace(s *discordgo.Session, i *discordgo.InteractionCreate) {
	race := GetCurrentRace(i.GuildID)
	if race == nil {
		slog.Warn("no race is planned", slog.String("guildID", i.GuildID), slog.String("memberID", i.Member.User.ID))
		disgomsg.NewResponse(disgomsg.WithContent("No race is planned")).SendEphemeral(s, i.Interaction)
		return
	}

	guildMember := guild.GetMember(i.GuildID, i.Member.User.ID).SetName(i.Member.User.Username, i.Member.Nick, i.Member.User.GlobalName)
	raceMember := getRaceMember(i.GuildID, guildMember)

	if _, err := race.addRaceParticipant(raceMember); err != nil {
		disgomsg.NewResponse(disgomsg.WithContent(unicode.FirstToUpper(err.Error()))).SendEphemeral(s, i.Interaction)
		return
	}

	slog.Debug("joined the race", slog.String("guildID", i.GuildID), slog.String("memberID", i.Member.User.ID))
	disgomsg.NewResponse(disgomsg.WithContent("You have joined the race")).SendEphemeral(s, i.Interaction)

	raceMessage(s, race, "join")
}

// raceStats returns a players race stats.
func raceStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	lang, err := language.Parse(string(i.Locale))
	if err != nil {
		lang = language.AmericanEnglish
	}
	p := message.NewPrinter(lang)

	// Update the member's name in the guild.
	guildMember := guild.GetMember(i.GuildID, i.Member.User.ID).SetName(i.Member.User.Username, i.Member.Nick, i.Member.User.GlobalName)
	raceMember := getRaceMember(i.GuildID, guildMember)

	var totalRaces float64
	if raceMember.TotalRaces == 0 {
		totalRaces = 1
	} else {
		totalRaces = float64(raceMember.TotalRaces)
	}

	var betsMade float64
	if raceMember.BetsMade == 0 {
		betsMade = 1
	} else {
		betsMade = float64(raceMember.BetsMade)
	}

	embeds := []*discordgo.MessageEmbed{
		{
			Type:  discordgo.EmbedTypeRich,
			Title: guildMember.Name,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "First",
					Value:  p.Sprintf("%d (%.0f%%)", raceMember.RacesWon, 100*float64(raceMember.RacesWon)/totalRaces),
					Inline: true,
				},
				{
					Name:   "Second",
					Value:  p.Sprintf("%d (%.0f%%)", raceMember.RacesPlaced, 100*float64(raceMember.RacesPlaced)/totalRaces),
					Inline: true,
				},
				{
					Name:   "Third",
					Value:  p.Sprintf("%d (%.0f%%)", raceMember.RacesShowed, 100*float64(raceMember.RacesShowed)/totalRaces),
					Inline: true,
				},
				{
					Name:   "Losses",
					Value:  p.Sprintf("%d (%.0f%%)", raceMember.RacesLost, 100*float64(raceMember.RacesLost)/totalRaces),
					Inline: true,
				},
				{
					Name:   "Earnings",
					Value:  p.Sprintf("%d", raceMember.TotalEarnings),
					Inline: true,
				},
				{
					Name:   "Bets Won",
					Value:  p.Sprintf("%d (%.0f%%)", raceMember.BetsWon, 100*float64(raceMember.BetsWon)/betsMade),
					Inline: true,
				},
				{
					Name:   "Bet Earnings",
					Value:  p.Sprintf("%d", raceMember.BetsEarnings),
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
		slog.Error("Unable to send the player stats to Discord",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.String("name", guildMember.Name),
			slog.Any("error", err),
		)
	}
}

// betOnRace processes a bet placed by a member on the race.
func betOnRace(s *discordgo.Session, i *discordgo.InteractionCreate) {
	race := GetCurrentRace(i.GuildID)
	if race == nil {
		slog.Warn("no race is planned", slog.String("guildID", i.GuildID), slog.String("memberID", i.Member.User.ID))
		disgomsg.NewResponse(disgomsg.WithContent("No race is planned")).SendEphemeral(s, i.Interaction)
		return
	}

	// Check to see if the member can place a bet
	err := raceBetChecks(race, i.Member.User.ID)
	if err != nil {
		slog.Error("unable to place bet", slog.String("guildID", i.GuildID), slog.String("memberID", i.Member.User.ID), slog.Any("error", err))
		disgomsg.NewResponse(disgomsg.WithContent(unicode.FirstToUpper(err.Error()))).SendEphemeral(s, i.Interaction)
		return
	}

	participant := race.getRaceParticipant(i.Member.User.ID)
	var betMember *RaceMember
	if (participant != nil) && (participant.Member != nil) {
		betMember = participant.Member
	} else {
		guildMember := guild.GetMember(i.GuildID, i.Member.User.ID).SetName(i.Member.User.Username, i.Member.Nick, i.Member.User.GlobalName)
		betMember = getRaceMember(i.GuildID, guildMember)
	}

	raceParticipant := getCurrentRaceParticipant(race, i.Interaction.MessageComponentData().CustomID)
	better := getRaceBetter(betMember, raceParticipant)
	if err = placeBet(race, better); err != nil {
		slog.Error("unable to place bet", slog.String("guildID", i.GuildID), slog.String("memberID", i.Member.User.ID), slog.Any("error", err))
		disgomsg.NewResponse(disgomsg.WithContent(fmt.Sprintf("Unable to place a bet, Error: %s", err.Error()))).SendEphemeral(s, i.Interaction)
		return
	}

	p := message.NewPrinter(language.AmericanEnglish)
	disgomsg.NewResponse(disgomsg.WithContent(p.Sprintf("You have placed a %d credit bet on %s", race.config.BetAmount, raceParticipant.Member.guildMember.Name))).SendEphemeral(s, i.Interaction)

	slog.Info("race bet placed", slog.String("guildID", i.GuildID), slog.String("memberID", i.Member.User.ID), slog.String("racer", raceParticipant.Member.guildMember.Name))
}

// createBetButtons returns the buttons for the racers, which may be used to
// bet on the various racers.
func createBetButtons(race *Race) []discordgo.ActionsRow {
	buttonsPerRow := 5
	rows := make([]discordgo.ActionsRow, 0, len(race.Racers)/buttonsPerRow)

	racersIncludedInButtons := 0
	for len(race.Racers) > racersIncludedInButtons {
		racersNotInButtons := len(race.Racers) - racersIncludedInButtons
		buttonsForNextRow := min(buttonsPerRow, racersNotInButtons)
		buttons := make([]discordgo.MessageComponent, 0, buttonsForNextRow)
		for i := range buttonsForNextRow {
			index := i + racersIncludedInButtons
			racer := race.Racers[index]
			button := discordgo.Button{
				Label:    race.Racers[index].Member.guildMember.Name,
				Style:    discordgo.PrimaryButton,
				CustomID: createBetButton(racer).label,
				Emoji:    nil,
			}
			buttons = append(buttons, button)
		}
		racersIncludedInButtons += buttonsForNextRow

		row := discordgo.ActionsRow{Components: buttons}
		rows = append(rows, row)
	}

	return rows
}

// createBetButton creates and returns a new race button for the racer, as well as
// registers the handlers for the button with Discord.
func createBetButton(rp *RaceParticipant) *raceButton {
	betButtonMutex.Lock()
	defer betButtonMutex.Unlock()

	// Get the race buttons for the guild
	buttons := betButtons[rp.Member.GuildID]
	if buttons == nil {
		buttons = make(map[string]*raceButton)
		betButtons[rp.Member.GuildID] = buttons
	}

	// Add a new button to the guild's button list
	label := fmt.Sprintf("%s:%s", rp.Member.GuildID, rp.Member.MemberID)
	button := buttons[label]
	if button != nil {
		return button
	}

	button = &raceButton{
		label: label,
		racer: rp,
	}
	buttons[button.label] = button

	// Register the component handler for the button
	bot.AddComponentHandler(button.label, betOnRace)
	slog.Debug("registered button component handler", slog.String("guildID", rp.Member.GuildID), slog.String("memberID", rp.Member.MemberID), slog.String("label", button.label))

	return button
}

// removeBetButtons removes the buttons for the current race and de-registers the
// handlers for all buttons in the race from Discord.
func removeBetButtons(race *Race) {
	betButtonMutex.Lock()
	defer betButtonMutex.Unlock()

	buttons := betButtons[race.GuildID]
	for key := range buttons {
		bot.RemoveComponentHandler(key)
		slog.Debug("removed race button component handler", slog.String("guildID", race.GuildID), slog.String("label", key))
	}
	betButtons[race.GuildID] = make(map[string]*raceButton)
}

// raceMessage sends the main command used to start and join the race. It also handles the case where
// the race begins, disabling the buttons to join the race.
func raceMessage(s *discordgo.Session, race *Race, action string) error {
	// handle edge case where a person joins before the race creation message is sent
	if race.interaction == nil {
		return errors.New("interaction is nil for the race")
	}
	p := message.NewPrinter(language.AmericanEnglish)

	racerNames := make([]string, 0, len(race.Racers))
	for _, racer := range race.Racers {
		racerNames = append(racerNames, racer.Member.guildMember.Name)
	}

	var msg string
	switch action {
	case "start", "join", "update":
		until := time.Until(race.RaceStartTime.Add(race.config.WaitToStart))
		msg = p.Sprintf(":triangular_flag_on_post: A race is starting! Click the button to join the race! :triangular_flag_on_post:\n\t\t\t\t\tThe race will begin in %s!", format.Duration(until))
	case "betting":
		until := time.Until(race.RaceStartTime.Add(race.config.WaitToStart + race.config.WaitForBets))
		msg = p.Sprintf(":triangular_flag_on_post: The racers have been set - betting is now open! :triangular_flag_on_post:\n\t\tYou have %s to place a %d credit bet!", format.Duration(until), race.config.BetAmount)
	case "started":
		msg = ":checkered_flag: The race is now in progress! :checkered_flag:"
	case "ended":
		msg = ":checkered_flag: The race has ended - lets find out the results. :checkered_flag:"
	case "cancelled":
		msg = "Not enough players entered the race, so it was cancelled."
	default:
		errMsg := fmt.Sprintf("Unrecognized action: %s", action)
		slog.Error(errMsg)
		return errors.New(errMsg)
	}

	embeds := []*discordgo.MessageEmbed{
		{
			Type:  discordgo.EmbedTypeRich,
			Title: "Race",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   msg,
					Inline: false,
				},
				{
					Name:   p.Sprintf("Racers (%d)", len(race.Racers)),
					Value:  strings.Join(racerNames, ", "),
					Inline: false,
				},
			},
		},
	}

	var err error
	switch action {
	case "start":
		components := []discordgo.MessageComponent{
			discordgo.ActionsRow{Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Join",
					Style:    discordgo.SuccessButton,
					CustomID: "join_race",
					Emoji:    nil,
				},
			}},
		}
		err = s.InteractionRespond(race.interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds:     embeds,
				Components: components,
			},
		})
	case "join":
		components := []discordgo.MessageComponent{
			discordgo.ActionsRow{Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Join",
					Style:    discordgo.SuccessButton,
					CustomID: "join_race",
					Emoji:    nil,
				},
			}},
		}
		_, err = s.InteractionResponseEdit(race.interaction.Interaction, &discordgo.WebhookEdit{
			Embeds:     &embeds,
			Components: &components,
		})
	case "update":
		_, err = s.InteractionResponseEdit(race.interaction.Interaction, &discordgo.WebhookEdit{
			Embeds: &embeds,
		})
	case "betting":
		var components []discordgo.MessageComponent
		rows := createBetButtons(race)
		for _, row := range rows {
			components = append(components, row)
		}
		_, err = s.InteractionResponseEdit(race.interaction.Interaction, &discordgo.WebhookEdit{
			Embeds:     &embeds,
			Components: &components,
		})
	default:
		_, err = s.InteractionResponseEdit(race.interaction.Interaction, &discordgo.WebhookEdit{
			Embeds:     &embeds,
			Components: &[]discordgo.MessageComponent{},
		})
	}

	return err
}

// Send the race so the guild members can watch it play out
func sendRaceLegs(s *discordgo.Session, race *Race) {
	channelID := race.interaction.ChannelID
	// Send the initial track
	track := getCurrentTrack(race.RaceLegs[0], race.config)
	m, err := s.ChannelMessageSend(channelID, fmt.Sprintf("%s\n", track))
	if err != nil {
		slog.Error("failed to send message at the start of the race", slog.String("guildID", race.GuildID), slog.String("channelID", channelID), slog.Any("error", err))
		return
	}

	slog.Debug("preparing to send race legs", slog.String("guildID", race.GuildID))
	for _, raceLeg := range race.RaceLegs {
		time.Sleep(2 * time.Second)
		track = getCurrentTrack(raceLeg, race.config)
		if _, err = s.ChannelMessageEdit(channelID, m.ID, fmt.Sprintf("%s\n", track)); err != nil {
			slog.Error("failed to update race message", slog.String("guildID", race.GuildID), slog.String("channelID", channelID), slog.Any("error", err))
		}
	}
}

// getCurrentTrack returns the current position of all racers on the track
func getCurrentTrack(raceLeg *RaceLeg, config *Config) string {
	var track strings.Builder
	for _, pos := range raceLeg.ParticipantPositions {
		name := pos.RaceParticipant.Member.guildMember.Name
		racer := pos.RaceParticipant.Racer

		position := max(0, pos.Position)

		start, end := unicode.SplitString(config.Track, position)
		currentTrackLine := start + racer.Emoji + end

		line := fmt.Sprintf("%s **%s %s** [%s]\n", config.EndingLine, currentTrackLine, config.StartingLine, name)
		track.WriteString(line)
	}
	return track.String()
}

// sendRaceResults sends the results of a race to the Discord server
func sendRaceResults(s *discordgo.Session, channelID string, race *Race) {
	p := message.NewPrinter(language.English)
	raceResults := make([]*discordgo.MessageEmbedField, 0, 4)

	racers := race.Racers

	results := race.RaceResult

	if results.Win != nil {
		raceParticipant := results.Win.Participant
		memberName := raceParticipant.Member.guildMember.Name
		raceResults = append(raceResults, &discordgo.MessageEmbedField{
			Name:   p.Sprintf(":first_place: %s", memberName),
			Value:  p.Sprintf("%s\n%.2fs\nPrize: %d", raceParticipant.Racer.Emoji, results.Win.RaceTime, results.Win.Winnings),
			Inline: true,
		})
	}

	if results.Place != nil {
		raceParticipant := results.Place.Participant
		memberName := raceParticipant.Member.guildMember.Name
		raceResults = append(raceResults, &discordgo.MessageEmbedField{
			Name:   p.Sprintf(":second_place: %s", memberName),
			Value:  p.Sprintf("%s\n%.2fs\nPrize: %d", raceParticipant.Racer.Emoji, results.Place.RaceTime, results.Place.Winnings),
			Inline: true,
		})
	}
	if results.Show != nil {
		raceParticipant := results.Show.Participant
		memberName := raceParticipant.Member.guildMember.Name
		raceResults = append(raceResults, &discordgo.MessageEmbedField{
			Name:   p.Sprintf(":third_place: %s", memberName),
			Value:  p.Sprintf("%s\n%.2fs\nPrize: %d", raceParticipant.Racer.Emoji, results.Show.RaceTime, results.Show.Winnings),
			Inline: true,
		})
	}

	betWinners := make([]string, 0, len(race.Betters))
	for _, bet := range race.Betters {
		if bet.Winnings > 0 {
			memberName := bet.Member.guildMember.Name
			betWinners = append(betWinners, memberName)
		}
	}
	var winners string
	if len(betWinners) > 0 {
		winners = strings.Join(betWinners, "\n")
	} else {
		winners = "No one guessed the winner."
	}
	betEarnings := race.config.BetAmount * len(racers)
	betResults := &discordgo.MessageEmbedField{
		Name:   p.Sprintf("Bet earnings of %d", betEarnings),
		Value:  winners,
		Inline: false,
	}
	raceResults = append(raceResults, betResults)
	embeds := []*discordgo.MessageEmbed{
		{
			Title:  "Race Results",
			Fields: raceResults,
		},
	}
	_, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds: embeds,
	})
	if err != nil {
		slog.Error("failed to send the racer results to the channel",
			slog.String("channelID", channelID),
			slog.Any("error", err),
		)
	}
}

// getRacer takes a custom button ID and returns the corresponding racer.
func getCurrentRaceParticipant(race *Race, customID string) *RaceParticipant {
	betButtonMutex.Lock()
	defer betButtonMutex.Unlock()

	buttons := betButtons[race.GuildID]
	button := buttons[customID]
	return button.racer
}
