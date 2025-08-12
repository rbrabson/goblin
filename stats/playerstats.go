package stats

import (
	"log/slog"
	"slices"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	millisToDays = 1000 * 60 * 60 * 24
)

var (
	statsLock = sync.Mutex{}
)

// PlayerStats holds the statistics of a player in a game.
type PlayerStats struct {
	ID                  primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID             string             `json:"guild_id" bson:"guild_id"`
	MemberID            string             `json:"member_id" bson:"member_id"`
	Game                string             `json:"game" bson:"game"`
	FirstPlayed         time.Time          `json:"first_played" bson:"first_played"`
	LastPlayed          time.Time          `json:"last_played" bson:"last_played"`
	NumberOfTimesPlayed int                `json:"number_of_times_played" bson:"number_of_times_played"`
}

// PlayerRetention holds the retention statistics of players in a game.
type PlayerRetention struct {
	InactivePlayers    int     `json:"inactive_players"`
	InactivePercentage float64 `json:"inactive_percentage"`
	ActivePlayers      int     `json:"active_players"`
	ActivePercentage   float64 `json:"active_percentage"`
}

// getPlayerStats retrieves the player statistics for a specific guild, member, and game.
// If the player stats do not exist, it creates a new PlayerStats instance.
func getPlayerStats(guildID string, memberID string, game string) *PlayerStats {
	ps, _ := readPlayerStats(guildID, memberID, game)
	if ps == nil {
		ps = newPlayerStats(guildID, memberID, game)
	}

	return ps
}

// newPlayerStats creates a new PlayerStats instance with the current date as FirstPlayed and LastPlayed.
func newPlayerStats(guildID string, memberID string, game string) *PlayerStats {
	today := today()
	ps := &PlayerStats{
		GuildID:             guildID,
		MemberID:            memberID,
		Game:                game,
		FirstPlayed:         today,
		LastPlayed:          time.Time{},
		NumberOfTimesPlayed: 0,
	}
	writePlayerStats(ps)
	return ps
}

// GetCurrentRanking returns the global rankings based on the current balance.
func getPlayerStatsForMostActiveMembers(guildID string, game string) []*PlayerStats {
	filter := bson.D{{Key: "guild_id", Value: guildID}, {Key: "game", Value: game}}
	sort := bson.D{{Key: "number_of_times_played", Value: -1}, {Key: "_id", Value: 1}}
	limit := int64(10)

	playerStats := readMultiplePlayerStats(guildID, filter, sort, limit)
	slices.SortFunc(playerStats, func(a, b *PlayerStats) int {
		switch {
		case a.NumberOfTimesPlayed > b.NumberOfTimesPlayed:
			return -1
		case a.NumberOfTimesPlayed < b.NumberOfTimesPlayed:
			return 1
		case a.MemberID < b.MemberID:
			return -1
		default:
			return 1
		}
	})

	return playerStats
}

// GetPlayerRetention finds players who played after a specific date but haven't played
// recently for the duration provided (i.e., players who became inactive)
func GetPlayerRetention(guildID string, game string, cuttoff time.Time, inactiveDuration time.Duration) (*PlayerRetention, error) {
	statsLock.Lock()
	defer statsLock.Unlock()

	slog.Debug("calculating player retention",
		slog.String("guild_id", guildID),
		slog.String("game", game),
		slog.Time("cuttoff", cuttoff),
		slog.Duration("inactive_duration", inactiveDuration),
	)

	today := today()
	inactiveDays := int(inactiveDuration.Hours()/24) + 1 // Convert duration to the number of days
	pipeline := make(mongo.Pipeline, 0, 7)

	if game == "" || game == "all" {
		// Stage 1: Match documents for the specific guild
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
			}},
		})
	} else {
		// Stage 1: Match documents for the specific guild and game
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
				{Key: "game", Value: game},
			}},
		})
	}

	// Stage 2: Group by member_id to get their first and last played dates
	pipeline = append(pipeline, bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$member_id"},
			{Key: "first_played", Value: bson.D{{Key: "$min", Value: "$first_played"}}},
			{Key: "last_played", Value: bson.D{{Key: "$max", Value: "$last_played"}}},
			{Key: "total_games", Value: bson.D{{Key: "$sum", Value: "$number_of_times_played"}}},
		}},
	})

	// Stage 3: Filter players who played after the specified date
	pipeline = append(pipeline, bson.D{
		{Key: "$match", Value: bson.D{
			{Key: "last_played", Value: bson.D{
				{Key: "$gte", Value: cuttoff},
			}},
		}},
	})

	// Stage 4: Add fields to calculate inactive status
	pipeline = append(pipeline, bson.D{
		{Key: "$addFields", Value: bson.D{
			{Key: "days_since_last_played", Value: bson.D{
				{Key: "$divide", Value: bson.A{
					bson.D{{Key: "$subtract", Value: bson.A{today, "$last_played"}}},
					millisToDays, // Convert milliseconds to days
				}},
			}},
		}},
	})

	// Stage 5: Categorize players as inactive or active
	pipeline = append(pipeline, bson.D{
		{Key: "$addFields", Value: bson.D{
			{Key: "is_active", Value: bson.D{
				{Key: "$cond", Value: bson.D{
					{Key: "if", Value: bson.D{
						{Key: "$lt", Value: bson.A{"$days_since_last_played", inactiveDays}},
					}},
					{Key: "then", Value: 1},
					{Key: "else", Value: 0},
				}},
			}},
			{Key: "is_inactive", Value: bson.D{
				{Key: "$cond", Value: bson.D{
					{Key: "if", Value: bson.D{
						{Key: "$gte", Value: bson.A{"$days_since_last_played", inactiveDays}},
					}},
					{Key: "then", Value: 1},
					{Key: "else", Value: 0},
				}},
			}},
		}},
	})

	// Stage 6: Group all players to calculate totals and percentages
	pipeline = append(pipeline, bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil}, // Group all documents together
			{Key: "total_players", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "active_players", Value: bson.D{{Key: "$sum", Value: "$is_active"}}},
			{Key: "inactive_players", Value: bson.D{{Key: "$sum", Value: "$is_inactive"}}},
		}},
	})

	// Stage 7: Calculate percentages
	pipeline = append(pipeline, bson.D{
		{Key: "$addFields", Value: bson.D{
			{Key: "inactive_percentage", Value: bson.D{
				{Key: "$multiply", Value: bson.A{
					bson.D{{Key: "$divide", Value: bson.A{"$inactive_players", "$total_players"}}},
					100,
				}},
			}},
			{Key: "active_percentage", Value: bson.D{
				{Key: "$multiply", Value: bson.A{
					bson.D{{Key: "$divide", Value: bson.A{"$active_players", "$total_players"}}},
					100,
				}},
			}},
		}},
	})

	docs, err := db.Aggregate(PlayerStatsCollection, pipeline)
	if err != nil {
		return nil, err
	}

	if len(docs) == 0 {
		return &PlayerRetention{
			InactivePlayers:    0,
			InactivePercentage: 0,
			ActivePlayers:      0,
			ActivePercentage:   0,
		}, nil
	}

	result := docs[0]
	retention := &PlayerRetention{
		InactivePlayers:    getInt(result["inactive_players"]), // Players who became inactive
		InactivePercentage: getFloat64(result["inactive_percentage"]),
		ActivePlayers:      getInt(result["active_players"]), // Players still active
		ActivePercentage:   getFloat64(result["active_percentage"]),
	}

	slog.Debug("player retention calculated",
		slog.Int("total_eligible_players", getInt(result["total_players"])),
		slog.Int("inactive_players", retention.InactivePlayers),
		slog.Float64("inactive_percentage", retention.InactivePercentage),
	)

	return retention, nil
}
