package race

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/internal/discmsg"
	"github.com/rbrabson/goblin/internal/format"
	"github.com/rbrabson/goblin/internal/unicode"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
)

var (
	racerButtonNames = []string{
		"race_bet_one",
		"race_bet_two",
		"race_bet_three",
		"race_bet_four",
		"race_bet_five",
		"race_bet_six",
		"race_bet_seven",
		"race_bet_eight",
		"race_bet_nine",
		"race_bet_ten",
	}

	componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"join_race":         joinRace,
		racerButtonNames[0]: betOnRace,
		racerButtonNames[1]: betOnRace,
		racerButtonNames[2]: betOnRace,
		racerButtonNames[3]: betOnRace,
		racerButtonNames[4]: betOnRace,
		racerButtonNames[5]: betOnRace,
		racerButtonNames[6]: betOnRace,
		racerButtonNames[7]: betOnRace,
		racerButtonNames[8]: betOnRace,
		racerButtonNames[9]: betOnRace,
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

// raceAdmin routes various `race-raceAdmin` subcommands to the appropriate handlers.
func raceAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> race.admin")
	defer log.Trace("<-- race.admin")

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "reset":
		resetRace(s, i)
	default:
		log.WithFields(log.Fields{"guild_id": i.GuildID, "user_id": i.Member.User.ID, "command": options[0].Name}).Error("unknown command")
		discmsg.SendEphemeralResponse(s, i, "Command is unknown")
	}
}

// race routes the various `race` subcommands to the appropriate handlers.
func race(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> race.race")
	defer log.Trace("<-- race.race")

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "start":
		startRace(s, i)
	case "stats":
		raceStats(s, i)
	default:
		discmsg.SendEphemeralResponse(s, i, "Command is unknown")
		log.WithFields(log.Fields{"guild_id": i.GuildID, "user_id": i.Member.User.ID, "command": options[0].Name}).Error("unknown command")
	}
}

// resetRace resets a hung race.
func resetRace(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> race.resetRace")
	defer log.Trace("<-- race.resetRace")

	ResetRace(i.GuildID)
	discmsg.SendEphemeralResponse(s, i, "Race has been reset")
}

// startRace starts a race that other members may join.
func startRace(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> race.startRace")
	defer log.Trace("<-- race.startRace")

	err := raceStartChecks(i.GuildID, i.Member.User.ID)
	if err != nil {
		discmsg.SendEphemeralResponse(s, i, err.Error())
		return
	}

	race := GetRace(i.GuildID)
	race.interaction = i
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
}

// waitForMembersToJoin waits until members join the race before proceeding
// to taking bets
func waitForMembersToJoin(s *discordgo.Session, race *Race) {
	log.Trace("--> waitForMembersToJoin")
	defer log.Trace("<-- waitForMembersToJoin")

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
			log.Error("Unable to update the time for the race message, error:", err)
		}
	}
}

// waitForBetsToBePlaced waits until bets are placed before starting the race.
func waitForBetsToBePlaced(s *discordgo.Session, race *Race) {
	log.Trace("--> waitForBetsToBePlaced")
	defer log.Trace("<-- waitForBetsToBePlaced")

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
			log.Error("Unable to update the time for the race message, error:", err)
		}
	}

}

// joinRace attempts to join a race that is getting ready to start.
func joinRace(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> race.joinRace")
	defer log.Trace("<-- race.joinRace")

	raceLock.Lock()
	defer raceLock.Unlock()

	race := currentRaces[i.GuildID]
	if race == nil {
		log.WithFields(log.Fields{"guild_id": i.GuildID}).Warn("no race is planned")
		discmsg.SendEphemeralResponse(s, i, "No race is planned")
		return
	}

	err := raceJoinChecks(race, i.Member.User.ID)
	if err != nil {
		discmsg.SendEphemeralResponse(s, i, err.Error())
		return
	}

	// All is good, add the member to the race
	raceMember := GetRaceMember(i.GuildID, i.Member.User.ID)
	race.addRaceParticipant(raceMember)
	log.WithFields(log.Fields{"guild_id": i.GuildID, "user_id": i.Member.User.ID}).Info("you have joined the race")
	discmsg.SendEphemeralResponse(s, i, "You have joined the race")

	err = raceMessage(s, race, "join")
	if err != nil {
		log.Error("Unable to update the race message, error:", err)
	}
}

// raceStats returns a players race stats.
func raceStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> race.raceStats")
	defer log.Trace("<-- race.raceStats")

	lang, err := language.Parse(string(i.Locale))
	if err != nil {
		lang = language.AmericanEnglish
	}
	p := discmsg.GetPrinter(lang)

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
		log.Error("Unable to send the player stats to Discord, error:", err)
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
		discmsg.SendEphemeralResponse(s, i, "No race is planned")
		return
	}

	err := raceBetChecks(race, i.Member.User.ID)
	if err != nil {
		discmsg.SendEphemeralResponse(s, i, err.Error())
		return
	}

	// Place the bet
	bankAccount := bank.GetAccount(i.GuildID, i.Member.User.ID)
	err = bankAccount.Withdraw(race.config.BetAmount)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "MemerID": i.Member.User.ID}).Error("unable to withdraw bet amount")
		discmsg.SendEphemeralResponse(s, i, "Insufficiient funds to place a bet")
		return
	}

	// All is good, so add the better to the race
	raceParticipant := getCurrentRaceParticipant(race, i.Interaction.MessageComponentData().CustomID)
	raceMember := GetRaceMember(i.GuildID, i.Member.User.ID)
	better := getRaceBetter(raceMember, raceParticipant)
	race.addBetter(better)
	p := discmsg.GetPrinter(language.AmericanEnglish)
	betMessage := p.Sprintf("You have placed a %d credit bet on %s", race.config.BetAmount, raceParticipant.Member.guildMember.Name)
	discmsg.SendEphemeralResponse(s, i, betMessage)

	log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": i.Member.User.ID, "racer": raceParticipant.Member.guildMember.Name}).Info("you have placed a bet")
}

// getRacerButtons returns the buttons for the racers, which may be used to
// bet on the various racers.
func getRacerButtons(race *Race) []discordgo.ActionsRow {
	log.Trace("--> getRacerButtons")
	defer log.Trace("<-- getRacerButtons")

	buttonsPerRow := 5
	rows := make([]discordgo.ActionsRow, 0, len(race.Racers)/buttonsPerRow)

	racersIncludedInButtons := 0
	for len(race.Racers) > racersIncludedInButtons {
		racersNotInButtons := len(race.Racers) - racersIncludedInButtons
		buttonsForNextRow := min(buttonsPerRow, racersNotInButtons)
		buttons := make([]discordgo.MessageComponent, 0, buttonsForNextRow)
		for j := 0; j < buttonsForNextRow; j++ {
			index := j + racersIncludedInButtons
			button := discordgo.Button{
				Label:    race.Racers[index].Member.guildMember.Name,
				Style:    discordgo.PrimaryButton,
				CustomID: racerButtonNames[index],
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

// raceMessage sends the main command used to start and join the race. It also handles the case where
// the race begins, disabling the buttons to join the race.
func raceMessage(s *discordgo.Session, race *Race, action string) error {
	log.Trace("--> raceMessage")
	defer log.Trace("<-- raceMessage")

	p := discmsg.GetPrinter(language.AmericanEnglish)

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
		rows := getRacerButtons(race)
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
	log.Trace("--> sendRace")
	defer log.Trace("<-- sendRace")

	channelID := race.interaction.ChannelID
	// Send the initial track
	track := getCurrentTrack(race.RaceLegs[0], race.config)
	message, err := s.ChannelMessageSend(channelID, fmt.Sprintf("%s\n", track))
	if err != nil {
		log.Error("Failed to send message at the start of the race, error:", err)
		return
	}

	log.Error("preparing to send race legs")
	for _, raceLeg := range race.RaceLegs {
		time.Sleep(2 * time.Second)
		track = getCurrentTrack(raceLeg, race.config)
		_, err = s.ChannelMessageEdit(channelID, message.ID, fmt.Sprintf("%s\n", track))
		if err != nil {
			log.Error("Failed to update race message, error:", err)
		}
	}
}

// getCurrentTrack returns the current position of all racers on the track
func getCurrentTrack(raceLeg *RaceLeg, config *Config) string {
	log.Trace("--> getCurrentTrack")
	defer log.Trace("<-- getCurrentTrack")

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
	log.Trace("--> sendRaceResults")
	defer log.Trace("<-- sendRaceResults")

	p := discmsg.GetPrinter(language.English)
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
		raceParticipant := results.Place.Participant
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
	log.Trace("--> getRacer")
	defer log.Trace("<-- getRacer")

	switch customID {
	case racerButtonNames[0]:
		return race.Racers[0]
	case racerButtonNames[1]:
		return race.Racers[1]
	case racerButtonNames[2]:
		return race.Racers[2]
	case racerButtonNames[3]:
		return race.Racers[3]
	case racerButtonNames[4]:
		return race.Racers[4]
	case racerButtonNames[5]:
		return race.Racers[5]
	case racerButtonNames[6]:
		return race.Racers[6]
	case racerButtonNames[7]:
		return race.Racers[7]
	case racerButtonNames[8]:
		return race.Racers[8]
	case racerButtonNames[9]:
		return race.Racers[9]
	case racerButtonNames[10]:
		return race.Racers[10]
	}

	log.Errorf("Invalid custom ID: %s", customID)
	return nil
}
