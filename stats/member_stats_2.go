package stats

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MemberStats2 struct {
	ID                  primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID             string             `json:"guild_id" bson:"guild_id"`
	MemberID            string             `json:"member_id" bson:"member_id"`
	Game                string             `json:"game" bson:"game"`
	FirstPlayed         time.Time          `json:"first_played" bson:"first_played"`
	LastPlayed          time.Time          `json:"last_played" bson:"last_played"`
	NumberOfTimesPlayed int                `json:"number_of_times_played" bson:"number_of_times_played"`
}
