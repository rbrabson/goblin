package race

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	RACE_CONFIG_COLLECTION = "race_configs"
	RACE_MEMBER_COLLECTION = "race_members"
	RACER_COLLECTION       = "race_avatars"
)

// readConfig loads the race configuration from the database. If it does not exist then
// a `nil` value is returned.
func readConfig(guildID string) *Config {
	filter := bson.M{"guild_id": guildID}
	var config Config
	err := db.FindOne(RACE_CONFIG_COLLECTION, filter, &config)
	if err != nil {
		sslog.Debug("race configuration not found in the database",
			slog.String("guildID", guildID),
			slog.String("error", err.Error()),
		)
		return nil
	}
	sslog.Debug("read race configuration from the database",
		slog.String("guildID", guildID),
		slog.String("config", config.Theme),
	)

	return &config
}

// writeConfig stores the race configuration in the database.
func writeConfig(config *Config) {
	var filter bson.M
	if config.ID != primitive.NilObjectID {
		filter = bson.M{"_id": config.ID}
	} else {
		filter = bson.M{"guild_id": config.GuildID}
	}
	err := db.UpdateOrInsert(RACE_CONFIG_COLLECTION, filter, config)
	if err != nil {
		sslog.Error("failed to write the race configuration to the database",
			slog.String("guildID", config.GuildID),
			slog.String("error", err.Error()),
		)
	}
}

// readConfig loads the race member from the database. If it does not exist then
// a `nil` value is returned.
func readRaceMember(guildID string, memberID string) *RaceMember {
	filter := bson.D{{Key: "guild_id", Value: guildID}, {Key: "member_id", Value: memberID}}
	var member RaceMember
	err := db.FindOne(RACE_MEMBER_COLLECTION, filter, &member)
	if err != nil {
		sslog.Debug("race member not found in the database",
			slog.String("guildID", guildID),
			slog.String("memberID", memberID),
			slog.String("error", err.Error()),
		)
		return nil
	}
	sslog.Debug("read race member from the database",
		slog.String("guildID", guildID),
		slog.String("memberID", memberID),
	)

	return &member
}

// Write creates or updates the race member in the database
func writeRaceMember(member *RaceMember) {
	var filter bson.M
	if member.ID != primitive.NilObjectID {
		filter = bson.M{"_id": member.ID}
	} else {
		filter = bson.M{"guild_id": member.GuildID, "member_id": member.MemberID}
	}
	db.UpdateOrInsert(RACE_MEMBER_COLLECTION, filter, member)
	sslog.Debug("write race member to the database",
		slog.String("guildID", member.GuildID),
		slog.String("memberID", member.MemberID),
	)
}

// readAllRaces loads the racers that may be used in racers that match the filter criteria.
func readAllRacers(filter bson.D) ([]*RaceAvatar, error) {
	var racers []*RaceAvatar
	sort := bson.D{{Key: "crew_size", Value: 1}}
	err := db.FindMany(RACER_COLLECTION, filter, &racers, sort, 0)
	if err != nil || len(racers) == 0 {
		sslog.Warn("unable to read racers",
			slog.String("error", err.Error()),
			"filter", filter,
		)
		if err != nil {
			return nil, err
		}
		return nil, ErrNoRacersFound
	}

	return racers, nil
}

// writeRacer creates or updates the target in the database.
func writeRacer(racer *RaceAvatar) {
	var filter bson.D
	if racer.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: racer.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: racer.GuildID}, {Key: "theme", Value: racer.Theme}, {Key: "emoji", Value: racer.Emoji}, {Key: "movement_speed", Value: racer.MovementSpeed}}
	}

	db.UpdateOrInsert(RACER_COLLECTION, filter, racer)
	sslog.Debug("create or update race avatar",
		slog.String("guildID", racer.GuildID),
		slog.String("theme", racer.Theme),
	)
}
