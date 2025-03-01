package heist

import (
	"github.com/rbrabson/goblin/guild"
	log "github.com/sirupsen/logrus"
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
	log.Trace("--> heist.readConfig")
	defer log.Trace("<-- heist.readConfig")

	filter := bson.M{"guild_id": guildID}
	var config Config
	err := db.FindOne(CONFIG_COLLECTION, filter, &config)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "error": err}).Debug("heist configuration not found in the database")
		return nil
	}
	log.WithFields(log.Fields{"guild": guildID}).Debug("read heist configuration from the database")

	return &config
}

// writeConfig stores the configuration in the database.
func writeConfig(config *Config) {
	log.Trace("--> heist.writeConfig")
	defer log.Trace("<-- heist.writeConfig")

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
func readMember(m *guild.Member) *HeistMember {
	log.Trace("--> heist.readMember")
	defer log.Trace("<-- heist.readMember")

	var heistMember HeistMember
	filter := bson.M{"guild_id": m.GuildID, "member_id": m.MemberID}
	err := db.FindOne(HEIST_MEMBER_COLLECTION, filter, &heistMember)
	if err != nil {
		log.WithFields(log.Fields{"guild": m.GuildID, "member": m.MemberID}).Debug("heist member not found in the database")
		return nil
	}
	heistMember.guildMember = m
	log.WithFields(log.Fields{"guild": heistMember.GuildID, "member": heistMember.MemberID}).Debug("read heist member from the database")

	return &heistMember
}

// Write creates or updates the heist member in the database
func writeMember(member *HeistMember) {
	log.Trace("--> heist.writeMember")
	defer log.Trace("<-- heist.writeMember")

	var filter bson.M
	if member.ID != primitive.NilObjectID {
		filter = bson.M{"_id": member.ID}
	} else {
		filter = bson.M{"guild_id": member.GuildID, "member_id": member.MemberID}
	}
	db.UpdateOrInsert(HEIST_MEMBER_COLLECTION, filter, member)
	log.WithFields(log.Fields{"guild": member.GuildID, "member": member.MemberID}).Debug("write heist member to the database")
}

// readAllTargets loads the targets that may be used in heists for all guilds
func readAllTargets(filter bson.D) ([]*Target, error) {
	log.Trace("--> heist.readAllTargets")
	defer log.Trace("<-- heist.readAllTargets")

	var targets []*Target
	sort := bson.D{{Key: "crew_size", Value: 1}}
	err := db.FindMany(TARGET_COLLECTION, filter, &targets, sort, 0)
	if err != nil {
		log.WithField("error", err).Error("unable to read targets")
		return nil, err
	}

	log.WithField("targets", targets).Trace("load targets")

	return targets, nil
}

// readTargets loads the targets that may be used in heists by the given guild
func readTargets(guildID string, theme string) ([]*Target, error) {
	log.Debug("--> heist.readTargets")
	defer log.Debug("<-- heist.readTargets")

	var targets []*Target
	filter := bson.D{{Key: "guild_id", Value: guildID}, {Key: "theme", Value: theme}}
	err := db.FindMany(TARGET_COLLECTION, filter, &targets, bson.D{}, 0)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "error": err}).Error("unable to read targets")
		return nil, err
	}

	log.WithFields(log.Fields{"guild": guildID, "targets": targets}).Trace("load targets")

	return targets, nil
}

// writeTarget writes the set of targets to the database. If they already exist, the are updated; otherwise, the set is created.
func writeTarget(target *Target) {
	log.Trace("--> heist.Target.writeTarget")
	defer log.Trace("<-- heist.Target.writeTarget")

	var filter bson.D
	if target.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: target.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: target.GuildID}, {Key: "target_id", Value: target.Name}}
	}

	db.UpdateOrInsert(TARGET_COLLECTION, filter, target)
	log.WithFields(log.Fields{"guild": target.GuildID, "target": target.Theme}).Debug("create or update target")
}

// readAllThemes loads all available themes for a guild
func readAllThemes(guildID string) ([]*Theme, error) {
	log.Trace("--> heist.readThemes")
	defer log.Trace("<-- heist.readThemes")

	var themes []*Theme
	filter := bson.D{{Key: "guild_id", Value: guildID}}
	err := db.FindMany(THEME_COLLECTION, filter, &themes, bson.D{}, 0)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "error": err}).Error("unable to read themes")
		return nil, err
	}

	log.WithFields(log.Fields{"guild": guildID, "themes": len(themes)}).Trace("read targets")

	return themes, nil
}

// readTheme loads the requested theme for a guild
func readTheme(guildID string, themeName string) (*Theme, error) {
	log.Trace("--> heist.readThemes")
	defer log.Trace("<-- heist.readThemes")

	var theme Theme
	filter := bson.D{{Key: "guild_id", Value: guildID}, {Key: "name", Value: themeName}}
	err := db.FindOne(THEME_COLLECTION, filter, &theme)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "themeName": themeName, "error": err}).Error("unable to read themes")
		return nil, err
	}

	log.WithFields(log.Fields{"guild": guildID, "theme": theme.Name}).Trace("read targets")

	return &theme, nil
}

// write creates or updates the theme in the database
func writeTheme(theme *Theme) {
	log.Trace("--> heist.writeTheme")
	defer log.Trace("<-- heist.writeTheme")

	var filter bson.M
	if theme.ID != primitive.NilObjectID {
		filter = bson.M{"_id": theme.ID}
	} else {
		filter = bson.M{"guild_id": theme.GuildID, "name": theme.Name}
	}
	db.UpdateOrInsert(THEME_COLLECTION, filter, theme)
	log.WithFields(log.Fields{"guild": theme.GuildID, "theme": theme.Name}).Debug("write theme to the database")
}
