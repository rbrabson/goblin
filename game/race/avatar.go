package race

import (
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"os"
	"path/filepath"

	"github.com/rbrabson/goblin/internal/emoji"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RaceAvatar represents a character that may be assigned to a member that partipates in a race
type RaceAvatar struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID       string             `json:"guild_id" bson:"guild_id"`
	Theme         string             `json:"theme" bson:"theme"`
	Emoji         string             `json:"emoji" bson:"emoji"`
	MovementSpeed string             `json:"movement_speed" bson:"movement_speed"`
}

// GetRaceAvatars returns the list of chracters that may be assigned to a member during a race.
func GetRaceAvatars(guildID string, themeName string) []*RaceAvatar {
	filter := bson.D{{Key: "guild_id", Value: guildID}, {Key: "theme", Value: themeName}}
	avatars, err := readAllRacers(filter)
	if err != nil {
		sslog.Warn("unable to read racers",
			slog.String("guildID", guildID),
			slog.String("theme", themeName),
			slog.String("error", err.Error()),
		)
		// If the list of characters does not exist, then create a new list.
		return readRaceAvatarsFromFile(guildID, themeName)
	}

	rand.Shuffle(len(avatars), func(i, j int) {
		avatars[i], avatars[j] = avatars[j], avatars[i]
	})

	sslog.Debug("read racers",
		slog.String("guildID", guildID),
		slog.String("theme", themeName),
		slog.Int("count", len(avatars)),
	)
	return avatars
}

// readRaceAvatarsFromFile reads the list of characters for the theme and guild from the database. If the list
// does not exist, then an error is returned.
func readRaceAvatarsFromFile(guildID string, themeName string) []*RaceAvatar {
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	configFileName := filepath.Join(configDir, "race", "avatars", themeName+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		sslog.Error("failed to read default race avatars",
			slog.String("guildID", guildID),
			slog.String("theme", themeName),
			slog.String("file", configFileName),
			slog.String("error", err.Error()),
		)
		return getDefaultRaceAvatars(guildID)
	}

	var avatars []*RaceAvatar
	err = json.Unmarshal(bytes, &avatars)
	if err != nil {
		sslog.Error("failed to unmarshal default race avatars",
			slog.String("guildID", guildID),
			slog.String("theme", themeName),
			slog.String("file", configFileName),
			slog.String("data", string(bytes)),
			slog.String("error", err.Error()))
		return getDefaultRaceAvatars(guildID)
	}

	for _, avatar := range avatars {
		avatar.GuildID = guildID
		avatar.Theme = themeName
		writeRacer(avatar)
	}

	sslog.Info("create new race avatars",
		slog.String("guildID", guildID),
		slog.String("theme", themeName),
		slog.Int("count", len(avatars)),
	)

	return avatars
}

// getDefaultRaceAvatars creates a new list of characters for the guild. The list is saved to
// the database.
func getDefaultRaceAvatars(guildID string) []*RaceAvatar {
	racers := []*RaceAvatar{
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Minion,
			MovementSpeed: "veryfast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Miner,
			MovementSpeed: "veryfast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Goblin,
			MovementSpeed: "veryfast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.BetaMinion,
			MovementSpeed: "veryfast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.WallBreaker,
			MovementSpeed: "fast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Valkrie,
			MovementSpeed: "fast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.SneakyArcher,
			MovementSpeed: "fast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.HogRider,
			MovementSpeed: "fast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.ArcherQueen,
			MovementSpeed: "fast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Archer,
			MovementSpeed: "fast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Barbarian,
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.CannonCart,
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Healer,
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Wizard,
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.BarbarianKing,
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.GrandWarden,
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.BattleMachine,
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Bomber,
			MovementSpeed: "abberant",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.DropShip,
			MovementSpeed: "abberant",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Balloon,
			MovementSpeed: "abberant",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.ElectroDragon,
			MovementSpeed: "predator",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Dragon,
			MovementSpeed: "predator",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.BattleBlimp,
			MovementSpeed: "predator",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.LavaHound,
			MovementSpeed: "predator",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.BabyDragon,
			MovementSpeed: "babydragon",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.RagedBarbarian,
			MovementSpeed: "special",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.SuperPEKKA,
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.PEKKA,
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Bowler,
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Witch,
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.WallWrecker,
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.NightWitch,
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Golem,
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Giant,
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.BoxerGiant,
			MovementSpeed: "slow",
		},
	}

	for _, racer := range racers {
		writeRacer(racer)
	}

	sslog.Info("created new racers",
		slog.String("guildID", guildID),
		slog.Int("count", len(racers)),
	)

	return racers
}

// calculateMovement calculates the distance a racer moves on a given turn
func (avatar *RaceAvatar) calculateMovement(currentTurn int) int {
	source := rand.NewPCG(rand.Uint64(), rand.Uint64())
	r := rand.New(source)
	switch avatar.MovementSpeed {
	case "veryfast":
		return r.IntN(8) * 2
	case "fast":
		return r.IntN(5) * 3
	case "slow":
		return (r.IntN(3) + 1) * 3
	case "steady":
		return 2 * 3
	case "abberant":
		chance := r.IntN(100)
		if chance >= 70 {
			return 5 * 3
		}
		return r.IntN(3) * 3
	case "predator":
		if currentTurn%2 != 0 {
			return 0
		} else {
			return (r.IntN(4) + 2) * 3
		}
	case "special":
		fallthrough
	default:
		switch currentTurn {
		case 1:
			return 7 * 3
		case 2:
			return 7 * 3
		default:
			return r.IntN(3) * 3
		}
	}
}

// String returns a string representation of the race avatar.
func (r *RaceAvatar) String() string {
	return r.Emoji
}
