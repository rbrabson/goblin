package heist

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	CONFIG_COLLECTION       = "heist_configs"
	HEIST_MEMBER_COLLECTION = "heist_members"
	TARGET_COLLECTION       = "heist_targets"
	THEME_COLLECTION        = "heist_themes"
)

// readConfig loads the heist configuration from the database. If it does not exist then
// a `nil` value is returned.
func readConfig(guildID string) *Config {
	filter := bson.M{"guild_id": guildID}
	var config Config
	err := db.FindOne(CONFIG_COLLECTION, filter, &config)
	if err != nil {
		slog.Debug("heist configuration not found in the database",
			slog.String("guildID", guildID),
			slog.Any("error", err),
		)
		return nil
	}
	slog.Debug("read heist configuration from the database",
		slog.String("guildID", guildID),
	)

	return &config
}

// writeConfig stores the configuration in the database.
func writeConfig(config *Config) {
	var filter bson.M
	if config.ID != primitive.NilObjectID {
		filter = bson.M{"_id": config.ID}
	} else {
		filter = bson.M{"guild_id": config.GuildID}
	}
	db.UpdateOrInsert(CONFIG_COLLECTION, filter, config)
}

// readMember loads the heist member from the database. If it does not exist then
// a `nil` value is returned.
func readMember(guildID string, memberID string) *HeistMember {
	var heistMember HeistMember
	filter := bson.M{"guild_id": guildID, "member_id": memberID}
	err := db.FindOne(HEIST_MEMBER_COLLECTION, filter, &heistMember)
	if err != nil {
		slog.Debug("heist member not found in the database",
			slog.String("guildID", guildID),
			slog.String("memberID", memberID),
			slog.Any("error", err),
		)
		return nil
	}
	slog.Debug("read heist member from the database",
		slog.String("guildID", guildID),
		slog.String("memberID", memberID),
	)

	return &heistMember
}

// Write creates or updates the heist member in the database
func writeMember(member *HeistMember) {
	var filter bson.M
	if member.ID != primitive.NilObjectID {
		filter = bson.M{"_id": member.ID}
	} else {
		filter = bson.M{"guild_id": member.GuildID, "member_id": member.MemberID}
	}
	db.UpdateOrInsert(HEIST_MEMBER_COLLECTION, filter, member)
	slog.Debug("write heist member to the database",
		slog.String("guildID", member.GuildID),
		slog.String("memberID", member.MemberID),
	)
}

// readAllTargets loads the targets that may be used in heists for all guilds
func readAllTargets(filter bson.D) ([]*Target, error) {
	var targets []*Target
	sort := bson.D{{Key: "crew_size", Value: 1}}
	err := db.FindMany(TARGET_COLLECTION, filter, &targets, sort, 0)
	if err != nil {
		slog.Error("unable to read targets",
			slog.Any("error", err),
			"filter", filter,
		)
		return nil, err
	}

	return targets, nil
}

// readTargets loads the targets that may be used in heists by the given guild
func readTargets(guildID string, theme string) ([]*Target, error) {
	var targets []*Target
	filter := bson.D{{Key: "guild_id", Value: guildID}, {Key: "theme", Value: theme}}
	err := db.FindMany(TARGET_COLLECTION, filter, &targets, bson.D{}, 0)
	if err != nil {
		slog.Error("unable to read targets",
			slog.String("guildID", guildID),
			slog.String("theme", theme),
			slog.Any("error", err),
		)
		return nil, err
	}

	return targets, nil
}

// writeTarget writes the set of targets to the database. If they already exist, the are updated; otherwise, the set is created.
func writeTarget(target *Target) {
	var filter bson.D
	if target.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: target.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: target.GuildID}, {Key: "target_id", Value: target.Name}}
	}

	db.UpdateOrInsert(TARGET_COLLECTION, filter, target)
	slog.Debug("create or update target",
		slog.String("guild", target.GuildID),
		slog.String("target", target.Name),
		slog.String("theme", target.Theme),
	)
}

// readAllThemes loads all available themes for a guild
func readAllThemes(guildID string) ([]*Theme, error) {
	var themes []*Theme
	filter := bson.D{{Key: "guild_id", Value: guildID}}
	err := db.FindMany(THEME_COLLECTION, filter, &themes, bson.D{}, 0)
	if err != nil {
		slog.Error("unable to read themes",
			slog.String("guildID", guildID),
			slog.Any("error", err),
		)
		return nil, err
	}

	return themes, nil
}

// readTheme loads the requested theme for a guild
func readTheme(guildID string, themeName string) (*Theme, error) {
	var theme Theme
	filter := bson.D{{Key: "guild_id", Value: guildID}, {Key: "name", Value: themeName}}
	err := db.FindOne(THEME_COLLECTION, filter, &theme)
	if err != nil {
		slog.Error("unable to read themes",
			slog.String("guild", guildID),
			slog.String("theme", themeName),
			slog.Any("error", err),
		)
		return nil, err
	}

	return &theme, nil
}

// write creates or updates the theme in the database
func writeTheme(theme *Theme) {
	var filter bson.M
	if theme.ID != primitive.NilObjectID {
		filter = bson.M{"_id": theme.ID}
	} else {
		filter = bson.M{"guild_id": theme.GuildID, "name": theme.Name}
	}
	db.UpdateOrInsert(THEME_COLLECTION, filter, theme)
	slog.Debug("write theme to the database",
		slog.String("guild", theme.GuildID),
		slog.String("theme", theme.Name),
	)
}
