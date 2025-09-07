package alt

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	AltCollection = "alt_ids"
)

// readAltID reads the alt ID from the database and returns the value, if it exists, or returns nil if the
// alt ID does not exist in the database.
func readAltID(guildID string, altID string) *AltID {
	filter := bson.M{
		"guild_id": guildID,
		"alt_id":   altID,
	}
	var alt AltID
	err := db.FindOne(AltCollection, filter, &alt)
	if err != nil {
		slog.Debug("alt ID not found in the database",
			slog.String("guildID", guildID),
			slog.String("altID", altID),
			slog.Any("error", err),
		)
		return nil
	}
	return &alt
}

// writeAltID creates or updates the alt ID data in the database.
func writeAltID(alt *AltID) error {
	var filter bson.M
	if alt.ID.IsZero() {
		filter = bson.M{
			"guild_id": alt.GuildID,
			"alt_id":   alt.AltID,
		}
	} else {
		filter = bson.M{"_id": alt.ID}
	}
	err := db.UpdateOrInsert(AltCollection, filter, alt)
	if err != nil {
		slog.Error("unable to create or update alt ID in the database",
			slog.String("guildID", alt.GuildID),
			slog.String("altID", alt.AltID),
			slog.Any("error", err),
		)
		return err
	}

	slog.Debug("write alt ID to the database",
		slog.String("guildID", alt.GuildID),
		slog.String("altID", alt.AltID),
	)
	return nil
}

func readAllAltIDs(guildID string, ownerID string) []*AltID {
	sort := bson.D{{Key: "alt_id", Value: 1}}
	var filter bson.M
	if ownerID != "" {
		filter = bson.M{
			"guild_id": guildID,
			"owner_id": ownerID,
		}
	} else {
		filter = bson.M{
			"guild_id": guildID,
		}
	}
	var alts []*AltID
	err := db.FindMany(AltCollection, filter, &alts, sort, 0)
	if err != nil {
		slog.Error("unable to fetch alt IDs from the database",
			slog.String("guildID", guildID),
			slog.Any("error", err),
		)
		return nil
	}
	return alts
}

// deleteAltID deletes the alt ID from the database.
func deleteAltID(guildID string, altID string) error {
	filter := bson.M{
		"guild_id": guildID,
		"alt_id":   altID,
	}
	err := db.Delete(AltCollection, filter)
	if err != nil {
		slog.Error("unable to delete alt ID from the database",
			slog.String("guildID", guildID),
			slog.String("altID", altID),
			slog.Any("error", err),
		)
		return err
	}

	slog.Debug("deleted alt ID from the database",
		slog.String("guildID", guildID),
		slog.String("altID", altID),
	)
	return nil
}
