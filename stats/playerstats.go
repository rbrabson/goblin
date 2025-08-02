package stats

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PlayerStats struct {
	ID                  primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID             string             `json:"guild_id" bson:"guild_id"`
	MemberID            string             `json:"member_id" bson:"member_id"`
	Game                string             `json:"game" bson:"game"`
	FirstPlayed         time.Time          `json:"first_played" bson:"first_played"`
	LastPlayed          time.Time          `json:"last_played" bson:"last_played"`
	NumberOfTimesPlayed int                `json:"number_of_times_played" bson:"number_of_times_played"`
}

type PlayerRetention struct {
	InactivePlayers    int     `json:"inactive_players"`
	InactivePercentage float64 `json:"inactive_percentage"`
	ActivePlayers      int     `json:"active_players"`
	ActivePercentage   float64 `json:"active_percentage"`
}

func GetPlayerRetention(startDate time.Time, duration time.Duration) (*PlayerRetention, error) {
	endDate := today().Add(-duration)

	// Pipeline to find the percentage of players who haven't played in the past month
	pipeline := mongo.Pipeline{
		// Stage 1: Match documents for the specific guild and game
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: "test_guild"},
				{Key: "game", Value: "test_game"},
				{Key: "first_played", Value: bson.D{ // Not working for some reason....
					{Key: "$gte", Value: startDate},
				}},
			}},
		},
		// Stage 2: Group by member_id and find their last activity day
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$member_id"},
				{Key: "last_played", Value: bson.D{{Key: "$max", Value: "$last_played"}}},
			}},
		},
		// Stage 3: Add a field to categorize players as active or inactive
		bson.D{
			{Key: "$addFields", Value: bson.D{
				{Key: "is_inactive", Value: bson.D{
					{Key: "$cond", Value: bson.D{
						{Key: "if", Value: bson.D{
							{Key: "$lt", Value: bson.A{"$last_played", endDate}},
						}},
						{Key: "then", Value: 1},
						{Key: "else", Value: 0},
					}},
				}},
			}},
		},
		// Stage 4: Group all players to calculate totals and percentages
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil}, // Group all documents together
				{Key: "total_players", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "inactive_players", Value: bson.D{{Key: "$sum", Value: "$is_inactive"}}},
				{Key: "active_players", Value: bson.D{
					{Key: "$sum", Value: bson.D{
						{Key: "$subtract", Value: bson.A{1, "$is_inactive"}},
					}},
				}},
			}},
		},
		// Stage 5: Calculate percentages
		bson.D{
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
		},
	}

	docs, err := db.Aggregate("member_stats", pipeline)
	if err != nil {
		return nil, err
	}

	if len(docs) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	result := docs[0]
	retention := &PlayerRetention{
		InactivePlayers:    getInt(result["inactive_players"]),
		InactivePercentage: getFloat64(result["inactive_percentage"]),
		ActivePlayers:      getInt(result["active_players"]),
		ActivePercentage:   getFloat64(result["active_percentage"]),
	}

	return retention, nil
}
