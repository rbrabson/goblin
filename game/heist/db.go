package heist

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	configCollection = "heist_configs"
	memberCollection = "heist_members"
	targetCollection = "heist_targets"
	themeCollection  = "heist_themes"
)

// readConfig loads the heist configuration from the database. If it does not exist, then
// a `nil` value is returned.
func readConfig(guildID string) *Config {
	filter := bson.M{"guild_id": guildID}
	var config Config
	err := db.FindOne(configCollection, filter, &config)
	if err != nil {
		slog.Debug("heist configuration not found in the database", slog.String("guildID", guildID), slog.Any("error", err))
		return nil
	}
	slog.Debug("read heist configuration from the database", slog.String("guildID", guildID))

	return &config
}

// writeConfig stores the configuration in the database.
func writeConfig(config *Config) {
	var filter bson.M
	if config.ID != bson.NilObjectID {
		filter = bson.M{"_id": config.ID}
	} else {
		filter = bson.M{"guild_id": config.GuildID}
	}
	if err := db.UpdateOrInsert(configCollection, filter, config); err != nil {
		slog.Error("error writing heist configuration to database", slog.String("guildID", config.GuildID), slog.Any("error", err))
	}
}

// readMember loads the heist member from the database. If it does not exist, then
// a `nil` value is returned.
func readMember(guildID string, memberID string) *HeistMember {
	var heistMember HeistMember
	filter := bson.M{"guild_id": guildID, "member_id": memberID}
	err := db.FindOne(memberCollection, filter, &heistMember)
	if err != nil {
		slog.Debug("heist member not found in the database", slog.String("guildID", guildID), slog.String("memberID", memberID), slog.Any("error", err))
		return nil
	}
	slog.Debug("read heist member from the database", slog.String("guildID", guildID), slog.String("memberID", memberID))

	return &heistMember
}

// Write creates or updates the heist member in the database
func writeMember(member *HeistMember) {
	var filter bson.M
	if member.ID != bson.NilObjectID {
		filter = bson.M{"_id": member.ID}
	} else {
		filter = bson.M{"guild_id": member.GuildID, "member_id": member.MemberID}
	}
	if err := db.UpdateOrInsert(memberCollection, filter, member); err != nil {
		slog.Error("error writing heist member to the database", slog.String("guildID", member.GuildID), slog.String("memberID", member.MemberID), slog.Any("error", err))
		return
	}
	slog.Debug("write heist member to the database", slog.String("guildID", member.GuildID), slog.String("memberID", member.MemberID))
}

// readAllTargets loads the targets that may be used in heists for all guilds
func readAllTargets(filter bson.M) ([]*Target, error) {
	var targets []*Target
	sort := bson.M{"crew_size": 1}
	err := db.FindMany(targetCollection, filter, &targets, sort, 0)
	if err != nil {
		slog.Error("unable to read targets", slog.Any("error", err), slog.Any("filter", filter))
		return nil, err
	}

	return targets, nil
}

// readTargets loads the targets that may be used in heists by the given guild
func readTargets(guildID string, theme string) ([]*Target, error) {
	var targets []*Target
	sort := bson.M{"crew_size": 1}
	filter := bson.M{"guild_id": guildID, "theme": theme}
	err := db.FindMany(targetCollection, filter, &targets, sort, 0)
	if err != nil {
		slog.Error("unable to read targets", slog.String("guildID", guildID), slog.String("theme", theme), slog.Any("error", err))
		return nil, err
	}

	return targets, nil
}

// writeTarget writes the set of targets to the database. If they already exist, they are updated; otherwise, the set is created.
func writeTarget(target *Target) {
	var filter bson.M
	if target.ID != bson.NilObjectID {
		filter = bson.M{"_id": target.ID}
	} else {
		filter = bson.M{"guild_id": target.GuildID, "target_id": target.Name}
	}

	if err := db.UpdateOrInsert(targetCollection, filter, target); err != nil {
		slog.Error("error writing target to database", slog.String("guildID", target.GuildID), slog.String("targetID", target.Name), slog.Any("error", err))
		return
	}
	slog.Debug("create or update target", slog.String("guild", target.GuildID), slog.String("target", target.Name), slog.String("theme", target.Theme))
}

// readAllThemes loads all available themes for a guild
func readAllThemes(guildID string) ([]*Theme, error) {
	var themes []*Theme
	filter := bson.M{"guild_id": guildID}
	err := db.FindMany(themeCollection, filter, &themes, bson.M{}, 0)
	if err != nil {
		slog.Error("unable to read themes", slog.String("guildID", guildID), slog.Any("error", err))
		return nil, err
	}

	return themes, nil
}

// readTheme loads the requested theme for a guild
func readTheme(guildID string, themeName string) (*Theme, error) {
	var theme Theme
	filter := bson.M{"guild_id": guildID, "name": themeName}
	err := db.FindOne(themeCollection, filter, &theme)
	if err != nil {
		slog.Error("unable to read theme", slog.String("guildID", guildID), slog.String("theme", themeName), slog.Any("error", err))
		return nil, err
	}

	return &theme, nil
}

// write creates or updates the theme in the database
func writeTheme(theme *Theme) {
	var filter bson.M
	if theme.ID != bson.NilObjectID {
		filter = bson.M{"_id": theme.ID}
	} else {
		filter = bson.M{"guild_id": theme.GuildID, "name": theme.Name}
	}
	if err := db.UpdateOrInsert(themeCollection, filter, theme); err != nil {
		slog.Error("error writing theme to the database", slog.String("guildID", theme.GuildID), slog.String("name", theme.Name), slog.Any("error", err))
		return
	}
	slog.Debug("write theme to the database", slog.String("guild", theme.GuildID), slog.String("theme", theme.Name))
}
