package slots

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	SlotsMemberCollection = "slots_members"
)

// readMember loads the slots member from the database. If it does not exist then
// a `nil` value is returned.
func readMember(guildID string, memberID string) *Member {
	var member Member
	filter := bson.M{"guild_id": guildID, "member_id": memberID}
	err := db.FindOne(SlotsMemberCollection, filter, &member)
	if err != nil {
		slog.Debug("slots member not found in the database",
			slog.String("guildID", guildID),
			slog.String("memberID", memberID),
			slog.Any("error", err),
		)
		return nil
	}
	slog.Debug("read slots member from the database",
		slog.String("guildID", guildID),
		slog.String("memberID", memberID),
	)

	return &member
}

// Write creates or updates the slots member in the database
func writeMember(member *Member) {
	var filter bson.M
	if member.ID != primitive.NilObjectID {
		filter = bson.M{"_id": member.ID}
	} else {
		filter = bson.M{"guild_id": member.GuildID, "member_id": member.MemberID}
	}
	if err := db.UpdateOrInsert(SlotsMemberCollection, filter, member); err != nil {
		slog.Error("error writing slots member to the database",
			slog.String("guildID", member.GuildID),
			slog.String("memberID", member.MemberID),
			slog.Any("error", err),
		)
	}
	slog.Debug("write slots member to the database",
		slog.String("guildID", member.GuildID),
		slog.String("memberID", member.MemberID),
	)
}

// PayoutAverages represents aggregated statistics across all slots members
type PayoutAverages struct {
	AverageTotalWins        float64 `bson:"average_total_wins"`
	AverageTotalLosses      float64 `bson:"average_total_losses"`
	AverageWinPercentage    float64 `bson:"average_win_percentage"`
	AverageLossPercentage   float64 `bson:"average_loss_percentage"`
	TotalWins               int64   `bson:"total_wins"`
	TotalLosses             int64   `bson:"total_losses"`
	TotalBet                int64   `bson:"total_bet"`
	TotalWon                int64   `bson:"total_won"`
	AverageReturns          float64 `bson:"average_returns"`
	AverageMaxWinningStreak float64 `bson:"average_max_winning_streak"`
	AverageMaxLosingStreak  float64 `bson:"average_max_losing_streak"`
}

// GetPayoutAverages uses an aggregation pipeline to calculate comprehensive payout statistics
// across all slots members in a guild
func GetPayoutAverages(guildID string) (*PayoutAverages, error) {
	pipeline := mongo.Pipeline{
		// Stage 1: Match documents for the specific guild
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
			}},
		},
		// Stage 2: Add computed fields
		bson.D{
			{Key: "$addFields", Value: bson.D{
				{Key: "total_games", Value: bson.D{
					{Key: "$add", Value: bson.A{"$total_wins", "$total_losses"}},
				}},
				{Key: "win_percentage", Value: bson.D{
					{Key: "$cond", Value: bson.D{
						{Key: "if", Value: bson.D{
							{Key: "$gt", Value: bson.A{
								bson.D{{Key: "$add", Value: bson.A{"$total_wins", "$total_losses"}}},
								0,
							}},
						}},
						{Key: "then", Value: bson.D{
							{Key: "$multiply", Value: bson.A{
								bson.D{{Key: "$divide", Value: bson.A{"$total_wins", bson.D{{Key: "$add", Value: bson.A{"$total_wins", "$total_losses"}}}}}},
								100,
							}},
						}},
						{Key: "else", Value: 0},
					}},
				}},
				{Key: "loss_percentage", Value: bson.D{
					{Key: "$cond", Value: bson.D{
						{Key: "if", Value: bson.D{
							{Key: "$gt", Value: bson.A{
								bson.D{{Key: "$add", Value: bson.A{"$total_wins", "$total_losses"}}},
								0,
							}},
						}},
						{Key: "then", Value: bson.D{
							{Key: "$multiply", Value: bson.A{
								bson.D{{Key: "$divide", Value: bson.A{"$total_losses", bson.D{{Key: "$add", Value: bson.A{"$total_wins", "$total_losses"}}}}}},
								100,
							}},
						}},
						{Key: "else", Value: 0},
					}},
				}},
				{Key: "returns", Value: bson.D{
					{Key: "$cond", Value: bson.D{
						{Key: "if", Value: bson.D{
							{Key: "$gt", Value: bson.A{"$total_bet", 0}},
						}},
						{Key: "then", Value: bson.D{
							{Key: "$multiply", Value: bson.A{
								bson.D{{Key: "$divide", Value: bson.A{"$total_winnings", "$total_bet"}}},
								100,
							}},
						}},
						{Key: "else", Value: 0},
					}},
				}},
			}},
		},
		// Stage 3: Group all documents and calculate averages
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil},
				{Key: "average_total_wins", Value: bson.D{
					{Key: "$avg", Value: "$total_wins"},
				}},
				{Key: "average_total_losses", Value: bson.D{
					{Key: "$avg", Value: "$total_losses"},
				}},
				{Key: "average_win_percentage", Value: bson.D{
					{Key: "$avg", Value: "$win_percentage"},
				}},
				{Key: "average_loss_percentage", Value: bson.D{
					{Key: "$avg", Value: "$loss_percentage"},
				}},
				{Key: "total_wins", Value: bson.D{
					{Key: "$sum", Value: "$total_wins"},
				}},
				{Key: "total_losses", Value: bson.D{
					{Key: "$sum", Value: "$total_losses"},
				}},
				{Key: "total_bet", Value: bson.D{
					{Key: "$sum", Value: "$total_bet"},
				}},
				{Key: "total_won", Value: bson.D{
					{Key: "$sum", Value: "$total_winnings"},
				}},
				{Key: "average_returns", Value: bson.D{
					{Key: "$avg", Value: "$returns"},
				}},
				{Key: "average_max_winning_streak", Value: bson.D{
					{Key: "$avg", Value: "$longest_win_streak"},
				}},
				{Key: "average_max_losing_streak", Value: bson.D{
					{Key: "$avg", Value: "$longest_losing_streak"},
				}},
			}},
		},
	}

	docs, err := db.Aggregate(SlotsMemberCollection, pipeline)
	if err != nil {
		slog.Error("failed to get payout averages",
			slog.String("guildID", guildID),
			slog.Any("error", err),
		)
		return nil, err
	}

	if len(docs) == 0 {
		slog.Debug("no slots data found for guild",
			slog.String("guildID", guildID),
		)
		// Return zero values if no data found
		return &PayoutAverages{}, nil
	}

	result := docs[0]

	// Extract values with proper type handling and defaults
	averages := &PayoutAverages{
		AverageTotalWins:        getFloatFromResult(result, "average_total_wins"),
		AverageTotalLosses:      getFloatFromResult(result, "average_total_losses"),
		AverageWinPercentage:    getFloatFromResult(result, "average_win_percentage"),
		AverageLossPercentage:   getFloatFromResult(result, "average_loss_percentage"),
		TotalWins:               getInt64FromResult(result, "total_wins"),
		TotalLosses:             getInt64FromResult(result, "total_losses"),
		TotalBet:                getInt64FromResult(result, "total_bet"),
		TotalWon:                getInt64FromResult(result, "total_won"),
		AverageReturns:          getFloatFromResult(result, "average_returns"),
		AverageMaxWinningStreak: getFloatFromResult(result, "average_max_winning_streak"),
		AverageMaxLosingStreak:  getFloatFromResult(result, "average_max_losing_streak"),
	}
	averages.AverageReturns = float64(averages.TotalWon) / float64(averages.TotalBet) * 100.0

	slog.Debug("retrieved payout averages",
		slog.String("guildID", guildID),
		slog.Any("averages", averages),
	)

	return averages, nil
}

// Helper function to safely extract float64 values from aggregation results
func getFloatFromResult(result primitive.M, key string) float64 {
	if val, ok := result[key]; ok && val != nil {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int64:
			return float64(v)
		case int32:
			return float64(v)
		case int:
			return float64(v)
		}
	}
	return 0.0
}

// Helper function to safely extract int64 values from aggregation results
func getInt64FromResult(result primitive.M, key string) int64 {
	if val, ok := result[key]; ok && val != nil {
		switch v := val.(type) {
		case int64:
			return v
		case int32:
			return int64(v)
		case int:
			return int64(v)
		case float64:
			return int64(v)
		case float32:
			return int64(v)
		}
	}
	return 0
}
