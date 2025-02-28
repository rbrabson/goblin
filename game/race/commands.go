package race

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/internal/discmsg"
	log "github.com/sirupsen/logrus"
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
		"race-admin": admin,
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

// admin routes various `race-admin` subcommands to the appropriate handlers.
func admin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> admin")
	defer log.Trace("<-- admin")

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
	log.Trace("--> race")
	defer log.Trace("<-- race")

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
	discmsg.SendEphemeralResponse(s, i, "reset not implemented")
	// TODO: implement
}

// startRace starts a race that other members may join.
func startRace(s *discordgo.Session, i *discordgo.InteractionCreate) {
	discmsg.SendEphemeralResponse(s, i, "start not implemented")
	// TODO: implement
}

// joinRace attempts to join a race that is getting ready to start.
func joinRace(s *discordgo.Session, i *discordgo.InteractionCreate) {
	discmsg.SendEphemeralResponse(s, i, "join not implemented")
	// TODO: implement
}

// raceStats returns a players race stats.
func raceStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	discmsg.SendEphemeralResponse(s, i, "stats not implemented")
	// TODO: implement
}

// betOnRace processes a bet placed by a member on the race.
func betOnRace(s *discordgo.Session, i *discordgo.InteractionCreate) {
	discmsg.SendEphemeralResponse(s, i, "resbetet not implemented")
	// TODO: implement
}
