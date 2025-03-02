package race

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/internal/discmsg"
	"github.com/rbrabson/goblin/internal/format"
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
		"race_bet_eleven",
	}

	componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"join_race":          joinRace,
		racerButtonNames[0]:  betOnRace,
		racerButtonNames[1]:  betOnRace,
		racerButtonNames[2]:  betOnRace,
		racerButtonNames[3]:  betOnRace,
		racerButtonNames[4]:  betOnRace,
		racerButtonNames[5]:  betOnRace,
		racerButtonNames[6]:  betOnRace,
		racerButtonNames[7]:  betOnRace,
		racerButtonNames[8]:  betOnRace,
		racerButtonNames[9]:  betOnRace,
		racerButtonNames[10]: betOnRace,
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

	// TODO: implement
	// Create a new race
	// add the starting person to the race
	// wait for others to join
	// if no one joined, cancel the race
	// run the race
	// return the results
	// end the race
	racers := GetRacers(i.GuildID, "clash")
	sb := strings.Builder{}
	sb.WriteString("start not implemented, emojis= ")
	for _, racer := range racers {
		sb.WriteString(racer.Emoji)
		sb.WriteString(" ")
	}

	discmsg.SendEphemeralResponse(s, i, sb.String())
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

	// TODO: bury all of this in a raceChecks() function call
	if len(race.RaceLegs) != 0 {
		log.WithFields(log.Fields{"guild_id": i.GuildID}).Warn("race is underway")
		discmsg.SendEphemeralResponse(s, i, "The race has already started, so you can't join.")
		return
	}

	config := GetConfig(i.GuildID)
	if config.MaxNumRacers == len(race.Racers) {
		p := discmsg.GetPrinter(language.AmericanEnglish)
		log.WithFields(log.Fields{"guild_id": i.GuildID, "maxRacers": config.MaxNumRacers}).Warn("max racers already entered")
		resp := p.Sprintf("You can't join the race, as there are already %d entered into the race.", config.MaxNumRacers)
		discmsg.SendEphemeralResponse(s, i, resp)
		return
	}

	racers := GetRacers(i.GuildID, config.Theme)
	raceMember := GetRaceMember(i.GuildID, i.Member.User.ID)
	raceParticipant := newRaceParticipcant(raceMember, racers)
	err := race.AddRacer(raceParticipant)
	if err != nil {
		log.WithFields(log.Fields{"guild_id": i.GuildID, "user_id": i.Member.User.ID}).Warn("already a member of the race")
		discmsg.SendEphemeralResponse(s, i, "You are already a member of the race")
		return
	}
	log.WithFields(log.Fields{"guild_id": i.GuildID, "user_id": i.Member.User.ID}).Info("you have joined the race")

	discmsg.SendEphemeralResponse(s, i, "You have joined the race")
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

	discmsg.SendEphemeralResponse(s, i, "resbetet not implemented")
	// TODO: implement
}

// waitOnRace waits for racers to join the race, or betters to bet on the race.
func waitOnRace() {
	log.Trace("--> race.waitOnRace")
	defer log.Trace("<-- race.waitOnRace")
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
			guildMember := guild.GetMember(race.GuildID, race.Racers[index].Member.MemberID)
			button := discordgo.Button{
				Label:    guildMember.Name,
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
func raceMessage(s *discordgo.Session, i *discordgo.InteractionCreate, action string, config *Config) error {
	log.Trace("--> raceMessage")
	defer log.Trace("<-- raceMessage")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	raceLock.Lock()
	race := currentRaces[i.GuildID]
	raceLock.Unlock()

	racerNames := make([]string, 0, len(race.Racers))
	for _, racer := range race.Racers {
		guildMember := guild.GetMember(i.GuildID, racer.Member.MemberID)
		racerNames = append(racerNames, guildMember.Name)
	}

	var msg string
	switch action {
	case "start", "join", "update":
		until := time.Until(race.RaceStartTime.Add(config.WaitToStart))
		msg = p.Sprintf(":triangular_flag_on_post: A race is starting! Click the button to join the race! :triangular_flag_on_post:\n\t\t\t\t\tThe race will begin in %s!", format.Duration(until))
	case "betting":
		until := time.Until(race.RaceStartTime.Add(config.WaitForBets))
		msg = p.Sprintf(":triangular_flag_on_post: The racers have been set - betting is now open! :triangular_flag_on_post:\n\t\tYou have %s to place a %d credit bet!", format.Duration(until), config.BetAmount)
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
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
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
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds:     &embeds,
			Components: &components,
		})
	case "update":
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &embeds,
		})
	case "betting":
		components := []discordgo.MessageComponent{}
		rows := getRacerButtons(race)
		for _, row := range rows {
			components = append(components, row)
		}
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds:     &embeds,
			Components: &components,
		})
	default:
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds:     &embeds,
			Components: &[]discordgo.MessageComponent{},
		})
	}

	return err
}

// sendRaceResults sends the results of a race to the Discord server
func sendRaceResults(s *discordgo.Session, channelID string, race *Race, config *Config) {
	log.Trace("--> sendRaceResults")
	defer log.Trace("<-- sendRaceResults")

	p := discmsg.GetPrinter(language.English)
	raceResults := make([]*discordgo.MessageEmbedField, 0, 4)

	racers := race.Racers

	results := race.RaceResult

	if results.Win == nil {
		raceParticipant := results.Win
		memberName := guild.GetMember(race.GuildID, raceParticipant.Member.MemberID).Name
		raceResults = append(raceResults, &discordgo.MessageEmbedField{
			Name:   p.Sprintf(":first_place: %s", memberName),
			Value:  p.Sprintf("%s\n%.2fs\nPrize: %d", raceParticipant.Racer.Emoji, raceParticipant.Racer.MovementSpeed, raceParticipant.Prize),
			Inline: true,
		})
	}

	if results.Place == nil {
		raceParticipant := results.Place
		memberName := guild.GetMember(race.GuildID, raceParticipant.Member.MemberID).Name
		raceResults = append(raceResults, &discordgo.MessageEmbedField{
			Name:   p.Sprintf(":first_place: %s", memberName),
			Value:  p.Sprintf("%s\n%.2fs\nPrize: %d", raceParticipant.Racer.Emoji, raceParticipant.Racer.MovementSpeed, raceParticipant.Prize),
			Inline: true,
		})
	}
	if results.Show == nil {
		raceParticipant := results.Place
		memberName := guild.GetMember(race.GuildID, raceParticipant.Member.MemberID).Name
		raceResults = append(raceResults, &discordgo.MessageEmbedField{
			Name:   p.Sprintf(":first_place: %s", memberName),
			Value:  p.Sprintf("%s\n%.2fs\nPrize: %d", raceParticipant.Racer.Emoji, raceParticipant.Racer.MovementSpeed, raceParticipant.Prize),
			Inline: true,
		})
	}

	betWinners := make([]string, 0, len(race.Betters))
	for _, bet := range race.Betters {
		if bet.Racer == results.Win {
			memberName := guild.GetMember(race.GuildID, bet.Member.MemberID).Name
			betWinners = append(betWinners, memberName)
		}
	}
	var winners string
	if len(betWinners) > 0 {
		winners = strings.Join(betWinners, "\n")
	} else {
		winners = "No one guessed the winner."
	}
	betEarnings := config.BetAmount * len(racers)
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
