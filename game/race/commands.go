package race

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/internal/format"
	"github.com/rbrabson/goblin/internal/unicode"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	raceButtons     = make(map[string]map[string]*raceButton)
	raceButtonMutex = sync.Mutex{}

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
		resp := disgomsg.Response{
			Content: "System is shutting down",
		}
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "reset":
		resetRace(s, i)
	default:
		log.WithFields(log.Fields{"guild_id": i.GuildID, "user_id": i.Member.User.ID, "command": options[0].Name}).Error("unknown command")
		resp := disgomsg.Response{
			Content: "Command is unknown",
		}
		resp.SendEphemeral(s, i.Interaction)
	}
}

// race routes the various `race` subcommands to the appropriate handlers.
func race(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		resp := disgomsg.Response{
			Content: "System is shutting down",
		}
		resp.Send(s, i.Interaction)
		return
	}

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "start":
		startRace(s, i)
	case "stats":
		raceStats(s, i)
	default:
		resp := disgomsg.Response{
			Content: "Command is unknown",
		}
		resp.Send(s, i.Interaction)
		log.WithFields(log.Fields{"guild_id": i.GuildID, "user_id": i.Member.User.ID, "command": options[0].Name}).Error("unknown command")
	}
}

// resetRace resets a hung race.
func resetRace(s *discordgo.Session, i *discordgo.InteractionCreate) {
	raceLock.Lock()
	defer raceLock.Unlock()

	ResetRace(i.GuildID)
	resp := disgomsg.Response{
		Content: "Race has been reset",
	}
	resp.SendEphemeral(s, i.Interaction)
}

// startRace starts a race that other members may join.
func startRace(s *discordgo.Session, i *discordgo.InteractionCreate) {
	raceLock.Lock()
	err := raceStartChecks(i.GuildID, i.Member.User.ID)
	if err != nil {
		raceLock.Unlock()
		resp := disgomsg.Response{
			Content: unicode.FirstToUpper(err.Error()),
		}
		resp.SendEphemeral(s, i.Interaction)
		return
	}
	race := GetRace(i.GuildID)
	race.interaction = i
	raceLock.Unlock()

	member := GetRaceMember(i.GuildID, i.Member.User.ID)
	race.addRaceParticipant(member)
	defer race.End()

	raceMessage(s, race, "start")
	log.WithFields(log.Fields{"guild_id": i.GuildID, "user_id": i.Member.User.ID}).Info("race started")
	waitForMembersToJoin(s, race)

	if len(race.Racers) < race.config.MinNumRacers {
		raceMessage(s, race, "cancelled")
		return
	}

	raceMessage(s, race, "betting")
	log.WithFields(log.Fields{"guild_id": i.GuildID, "racers": len(race.Racers)}).Info("waiting for bets")
	waitForBetsToBePlaced(s, race)

	raceMessage(s, race, "started")
	log.WithFields(log.Fields{"guild_id": i.GuildID, "betsPlaced": len(race.Betters)}).Info("race starting")

	race.RunRace(len([]rune(race.config.Track)))
	sendRace(s, race)

	raceMessage(s, race, "ended")
	log.WithFields(log.Fields{"guild_id": i.GuildID}).Info("race ended")

	sendRaceResults(s, i.ChannelID, race)
	removeRaceButtons(race)
}

// waitForMembersToJoin waits until members join the race before proceeding
// to taking bets
func waitForMembersToJoin(s *discordgo.Session, race *Race) {
	startTime := time.Now().Add(race.config.WaitToStart)
	for time.Now().Before(startTime) {
		maximumWait := time.Until(startTime)
		timeToWait := min(maximumWait, 5*time.Second)
		if timeToWait < 0 {
			break
		}
		time.Sleep(timeToWait)
		err := raceMessage(s, race, "update")
		if err != nil {
			log.WithError(err).Error("Unable to update the time for the race message")
		}
	}
}

// waitForBetsToBePlaced waits until bets are placed before starting the race.
func waitForBetsToBePlaced(s *discordgo.Session, race *Race) {
	betEndTime := time.Now().Add(race.config.WaitForBets)
	for time.Now().Before(betEndTime) {
		maximumWait := time.Until(betEndTime)
		timeToWait := min(maximumWait, 5*time.Second)
		if timeToWait < 0 {
			break
		}
		time.Sleep(timeToWait)
		err := raceMessage(s, race, "betting")
		if err != nil {
			log.WithError(err).Error("unable to update the time for the race message")
		}
	}
}

// joinRace attempts to join a race that is getting ready to start.
func joinRace(s *discordgo.Session, i *discordgo.InteractionCreate) {
	raceLock.Lock()
	defer raceLock.Unlock()

	race := currentRaces[i.GuildID]
	if race == nil {
		log.WithFields(log.Fields{"guild_id": i.GuildID}).Warn("no race is planned")
		resp := disgomsg.Response{
			Content: "No race is planned",
		}
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	err := raceJoinChecks(race, i.Member.User.ID)
	if err != nil {
		resp := disgomsg.Response{
			Content: unicode.FirstToUpper(err.Error()),
		}
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	// All is good, add the member to the race
	raceMember := GetRaceMember(i.GuildID, i.Member.User.ID)
	raceMember.guildMember.SetName(i.Member.User.Username, i.Member.DisplayName())
	race.addRaceParticipant(raceMember)
	log.WithFields(log.Fields{"guild_id": i.GuildID, "user_id": i.Member.User.ID}).Info("joined the race")
	resp := disgomsg.Response{
		Content: "You have joined the race",
	}
	err = resp.SendEphemeral(s, i.Interaction)
	if err != nil {
		log.WithFields(log.Fields{"guild_id": i.GuildID, "user_id": i.Member.User.ID}).WithError(err).Error("unable to send joined race message")
	}

	err = raceMessage(s, race, "join")
	if err != nil {
		log.WithError(err).Error("Unable to update the race message")
	}
}

// raceStats returns a players race stats.
func raceStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	lang, err := language.Parse(string(i.Locale))
	if err != nil {
		lang = language.AmericanEnglish
	}
	p := message.NewPrinter(lang)

	// Update the member's name in the guild.
	guildMember := guild.GetMember(i.GuildID, i.Member.User.ID).SetName(i.Member.User.Username, i.Member.DisplayName())

	raceMember := GetRaceMember(i.GuildID, i.Member.User.ID)

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
		log.WithError(err).Error("Unable to send the player stats to Discord")
	}
}

// betOnRace processes a bet placed by a member on the race.
func betOnRace(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("---> race.betOnRace")
	defer log.Trace("<--- race.betOnRace")

	raceLock.Lock()
	defer raceLock.Unlock()

	race := currentRaces[i.GuildID]
	if race == nil {
		log.WithFields(log.Fields{"guild_id": i.GuildID}).Warn("no race is planned")
		resp := disgomsg.Response{
			Content: "No race is planned",
		}
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	// Check to see if the member can place a bet
	err := raceBetChecks(race, i.Member.User.ID)
	if err != nil {
		resp := disgomsg.Response{
			Content: unicode.FirstToUpper(err.Error()),
		}
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	// Try to place the bet
	participant := race.getRaceParticipant(i.Member.User.ID)
	var raceMember *RaceMember
	if (participant != nil) && (participant.Member != nil) {
		raceMember = participant.Member
	} else {
		raceMember = GetRaceMember(i.GuildID, i.Member.User.ID)
	}
	raceParticipant := getCurrentRaceParticipant(race, i.Interaction.MessageComponentData().CustomID)
	err = raceMember.PlaceBet(race.config.BetAmount)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": i.Member.User.ID}).Error("unable to withdraw bet amount")
		resp := disgomsg.Response{

			Content: "Insufficiient funds in your bank account to place a bet",
		}
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	// Add to the list of betters
	raceMember.guildMember.SetName(i.Member.User.Username, i.Member.DisplayName())
	better := getRaceBetter(raceMember, raceParticipant)
	race.addBetter(better)
	p := message.NewPrinter(language.AmericanEnglish)
	resp := disgomsg.Response{
		Content: p.Sprintf("You have placed a %d credit bet on %s", race.config.BetAmount, raceParticipant.Member.guildMember.Name),
	}
	resp.SendEphemeral(s, i.Interaction)

	log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": i.Member.User.ID, "racer": raceParticipant.Member.guildMember.Name}).Info("you have placed a bet")
}

// getRaceButtons returns the buttons for the racers, which may be used to
// bet on the various racers.
func getRaceButtons(race *Race) []discordgo.ActionsRow {
	buttonsPerRow := 5
	rows := make([]discordgo.ActionsRow, 0, len(race.Racers)/buttonsPerRow)

	racersIncludedInButtons := 0
	for len(race.Racers) > racersIncludedInButtons {
		racersNotInButtons := len(race.Racers) - racersIncludedInButtons
		buttonsForNextRow := min(buttonsPerRow, racersNotInButtons)
		buttons := make([]discordgo.MessageComponent, 0, buttonsForNextRow)
		for j := 0; j < buttonsForNextRow; j++ {
			index := j + racersIncludedInButtons
			racer := race.Racers[index]
			button := discordgo.Button{
				Label:    race.Racers[index].Member.guildMember.Name,
				Style:    discordgo.PrimaryButton,
				CustomID: getRaceButton(racer).label,
				Emoji:    nil,
			}
			buttons = append(buttons, button)
		}
		racersIncludedInButtons += buttonsForNextRow

		row := discordgo.ActionsRow{Components: buttons}
		rows = append(rows, row)
		log.WithFields(log.Fields{
			"numRacers": len(race.Racers),
			"buttons":   len(buttons),
			"row":       len(rows),
		}).Trace("Race Buttons")
	}

	return rows
}

// getRaceButton creates and returns a new race button for the racer, as well as
// registers the handlers for the button with Discord.
func getRaceButton(rp *RaceParticipant) *raceButton {
	raceButtonMutex.Lock()
	defer raceButtonMutex.Unlock()

	// Get the race buttons for the guild
	buttons := raceButtons[rp.Member.GuildID]
	if buttons == nil {
		buttons = make(map[string]*raceButton)
		raceButtons[rp.Member.GuildID] = buttons
		log.WithFields(log.Fields{"guild_id": rp.Member.GuildID}).Trace("created new button list")
	}

	// Add a new button to the guild's button list
	label := fmt.Sprintf("%s:%s", rp.Member.GuildID, rp.Member.MemberID)
	button := buttons[label]
	if button != nil {
		log.WithFields(log.Fields{"guild_id": rp.Member.GuildID, "member_id": rp.Member.MemberID, "label": label}).Trace("found existing button")
		return button
	}

	button = &raceButton{
		label: label,
		racer: rp,
	}
	buttons[button.label] = button

	log.WithFields(log.Fields{"guild_id": rp.Member.GuildID, "member_id": rp.Member.MemberID, "lable": button.label}).Trace("created new button")

	// Register the component handler for the button
	bot.AddComponentHandler(button.label, betOnRace)
	log.WithFields(log.Fields{"guild_id": rp.Member.GuildID, "member_id": rp.Member.MemberID, "label": button.label}).Debug("registered button component handler")

	return button
}

// removeRaceButtons removes the buttons for the current race and de-registers the
// handlers for all buttons in the race from Discord.
func removeRaceButtons(race *Race) {
	raceButtonMutex.Lock()
	defer raceButtonMutex.Unlock()

	buttons := raceButtons[race.GuildID]
	for key := range buttons {
		bot.RemoveComponentHandler(key)
		log.WithFields(log.Fields{"guild_id": race.GuildID, "label": key}).Debug("removed button component handler")
	}
	raceButtons[race.GuildID] = make(map[string]*raceButton)
	log.WithFields(log.Fields{"guild_id": race.GuildID}).Trace("removed all buttons for the guild")
}

// raceMessage sends the main command used to start and join the race. It also handles the case where
// the race begins, disabling the buttons to join the race.
func raceMessage(s *discordgo.Session, race *Race, action string) error {
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
		log.Error(errMsg)
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
		components := []discordgo.MessageComponent{}
		rows := getRaceButtons(race)
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
func sendRace(s *discordgo.Session, race *Race) {
	channelID := race.interaction.ChannelID
	// Send the initial track
	track := getCurrentTrack(race.RaceLegs[0], race.config)
	message, err := s.ChannelMessageSend(channelID, fmt.Sprintf("%s\n", track))
	if err != nil {
		log.WithError(err).Error("failed to send message at the start of the race")
		return
	}

	log.Debug("preparing to send race legs")
	for _, raceLeg := range race.RaceLegs {
		time.Sleep(2 * time.Second)
		track = getCurrentTrack(raceLeg, race.config)
		_, err = s.ChannelMessageEdit(channelID, message.ID, fmt.Sprintf("%s\n", track))
		if err != nil {
			log.WithError(err).Error("Failed to update race message")
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
	s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds: embeds,
	})
}

// getRacer takes a custom button ID and returns the corresponding racer.
func getCurrentRaceParticipant(race *Race, customID string) *RaceParticipant {
	log.WithFields(log.Fields{"guild_id": race.GuildID, "customID": customID}).Trace("getting race participant for button")

	raceButtonMutex.Lock()
	defer raceButtonMutex.Unlock()

	buttons := raceButtons[race.GuildID]
	button := buttons[customID]
	return button.racer
}
