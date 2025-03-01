package race

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/internal/discmsg"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
)

var (
	componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"join_race":       joinRace,
		"race_bet_one":    betOnRace,
		"race_bet_two":    betOnRace,
		"race_bet_three":  betOnRace,
		"race_bet_four":   betOnRace,
		"race_bet_five":   betOnRace,
		"race_bet_six":    betOnRace,
		"race_bet_seven":  betOnRace,
		"race_bet_eight":  betOnRace,
		"race_bet_nine":   betOnRace,
		"race_bet_ten":    betOnRace,
		"race_bet_eleven": betOnRace,
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

	discmsg.SendEphemeralResponse(s, i, "start not implemented")
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
