package race

import (
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"os"
	"path/filepath"

	"github.com/rbrabson/goblin/discord"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Avatar represents a character that may be assigned to a member that partipates in a race
type Avatar struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID       string             `json:"guild_id" bson:"guild_id"`
	Theme         string             `json:"theme" bson:"theme"`
	Emoji         string             `json:"emoji" bson:"emoji"`
	MovementSpeed string             `json:"movement_speed" bson:"movement_speed"`
}

// GetRaceAvatars returns the list of chracters that may be assigned to a member during a race.
func GetRaceAvatars(guildID string, themeName string) []*Avatar {
	filter := bson.D{{Key: "guild_id", Value: guildID}, {Key: "theme", Value: themeName}}
	avatars, err := readAllRacers(filter)
	if err != nil {
		slog.Warn("unable to read racers",
			slog.String("guildID", guildID),
			slog.String("theme", themeName),
			slog.Any("error", err),
		)
		// If the list of characters does not exist, then create a new list.
		return readRaceAvatarsFromFile(guildID, themeName)
	}

	rand.Shuffle(len(avatars), func(i, j int) {
		avatars[i], avatars[j] = avatars[j], avatars[i]
	})

	slog.Debug("read racers",
		slog.String("guildID", guildID),
		slog.String("theme", themeName),
		slog.Int("count", len(avatars)),
	)
	return avatars
}

// readRaceAvatarsFromFile reads the list of characters for the theme and guild from the database. If the list
// does not exist, then an error is returned.
func readRaceAvatarsFromFile(guildID string, themeName string) []*Avatar {
	configFileName := filepath.Join(discord.DISCORD_CONFIG_DIR, "race", "avatars", themeName+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		slog.Error("failed to read default race avatars",
			slog.String("guildID", guildID),
			slog.String("theme", themeName),
			slog.String("file", configFileName),
			slog.Any("error", err),
		)
	}

	var avatars []*Avatar
	err = json.Unmarshal(bytes, &avatars)
	if err != nil {
		slog.Error("failed to unmarshal default race avatars",
			slog.String("guildID", guildID),
			slog.String("theme", themeName),
			slog.String("file", configFileName),
			slog.String("data", string(bytes)),
			slog.Any("error", err))
	}

	for _, avatar := range avatars {
		avatar.GuildID = guildID
		avatar.Theme = themeName
		writeRacer(avatar)
	}

	slog.Debug("create new race avatars",
		slog.String("guildID", guildID),
		slog.String("theme", themeName),
		slog.Int("count", len(avatars)),
	)

	return avatars
}

// calculateMovement calculates the distance a racer moves on a given turn
func (avatar *Avatar) calculateMovement(currentTurn int) int {
	config := GetConfig(avatar.GuildID)
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
	case "special", "babydragon":
		fallthrough
	default:
		switch currentTurn {
		case 1:
			return 7 * 3
		case 2:
			return 7 * 3
		default:
			movement := (r.IntN(3) * 3)
			if rand.IntN(100) < config.BabyDragonBuffPercent {
				movement += 1
			}
			return movement
		}
	}
}

// String returns a string representation of the race avatar.
func (avatar *Avatar) String() string {
	return avatar.Emoji
}
