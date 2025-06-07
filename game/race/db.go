package race

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	RaceConfigCollection = "race_configs"
	RaceMemberCollection = "race_members"
	RacerCollection      = "race_avatars"
)

// readConfig loads the race configuration from the database. If it does not exist then
// a `nil` value is returned.
func readConfig(guildID string) *Config {
	filter := bson.M{"guild_id": guildID}
	var config Config
	err := db.FindOne(RaceConfigCollection, filter, &config)
	if err != nil {
		slog.Debug("race configuration not found in the database",
			slog.String("guildID", guildID),
			slog.Any("error", err),
		)
		return nil
	}
	slog.Debug("read race configuration from the database",
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
	err := db.UpdateOrInsert(RaceConfigCollection, filter, config)
	if err != nil {
		slog.Error("failed to write the race configuration to the database",
			slog.String("guildID", config.GuildID),
			slog.Any("error", err),
		)
	}
}

// readConfig loads the race member from the database. If it does not exist then
// a `nil` value is returned.
func readRaceMember(guildID string, memberID string) *RaceMember {
	filter := bson.D{{Key: "guild_id", Value: guildID}, {Key: "member_id", Value: memberID}}
	var member RaceMember
	err := db.FindOne(RaceMemberCollection, filter, &member)
	if err != nil {
		slog.Debug("race member not found in the database",
			slog.String("guildID", guildID),
			slog.String("memberID", memberID),
			slog.Any("error", err),
		)
		return nil
	}
	slog.Debug("read race member from the database",
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
	if err := db.UpdateOrInsert(RaceMemberCollection, filter, member); err != nil {
		slog.Error("failed to write the race member to the database",
			slog.String("guildID", member.GuildID),
			slog.String("memberID", member.MemberID),
			slog.Any("error", err),
		)
	}
	slog.Info("write race member to the database",
		slog.String("guildID", member.GuildID),
		slog.String("memberID", member.MemberID),
	)
}

// readAllRaces loads the racers that may be used in racers that match the filter criteria.
func readAllRacers(filter bson.D) ([]*Avatar, error) {
	var racers []*Avatar
	sort := bson.D{{Key: "crew_size", Value: 1}}
	err := db.FindMany(RacerCollection, filter, &racers, sort, 0)
	if err != nil || len(racers) == 0 {
		slog.Warn("unable to read racers",
			slog.Any("error", err),
			slog.Any("filter", filter),
		)
		if err != nil {
			return nil, err
		}
		return nil, ErrNoRacersFound
	}

	return racers, nil
}

// writeRacer creates or updates the target in the database.
func writeRacer(racer *Avatar) {
	var filter bson.D
	if racer.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: racer.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: racer.GuildID}, {Key: "theme", Value: racer.Theme}, {Key: "emoji", Value: racer.Emoji}, {Key: "movement_speed", Value: racer.MovementSpeed}}
	}

	if err := db.UpdateOrInsert(RacerCollection, filter, racer); err != nil {
		slog.Error("failed to write the racer to the database",
			slog.String("guildID", racer.GuildID),
			slog.Any("error", err),
		)
	}
	slog.Info("create or update race avatar",
		slog.String("guildID", racer.GuildID),
		slog.String("theme", racer.Theme),
	)
}
