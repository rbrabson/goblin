package alt

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AltID represents an alternate ID associated with a guild member.
type AltID struct {
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID string             `json:"guild_id" bson:"guild_id"`
	OwnerID string             `json:"owner_id" bson:"owner_id"`
	AltID   string             `json:"alt_id" bson:"alt_id"`
}

// GetAltID retrieves an existing alt ID from the database, or creates a new one if it does not exist.
func GetAltID(guildID string, ownerID string, altID string) *AltID {
	alt := readAltID(guildID, altID)
	if alt != nil {
		return alt
	}
	return newAltID(guildID, ownerID, altID)
}

// newAltID creates a new alt ID and writes it to the database.
func newAltID(guildID string, ownerID string, altID string) *AltID {
	alt := &AltID{
		GuildID: guildID,
		AltID:   altID,
		OwnerID: ownerID,
	}
	writeAltID(alt)
	return alt
}

// GetAllAltIDs retrieves all alt IDs for a given guild from the database.
func GetAllAltIDs(guildID string) []*AltID {
	altIDs := readAllAltIDs(guildID, "")
	return altIDs
}

// IsAltID checks if the given alt ID exists in the database for the specified guild.
func IsAltID(guildID string, altID string) bool {
	alt := readAltID(guildID, altID)
	return alt != nil
}

// GetIDs returns a slice of alt ID strings from a slice of AltID structs.
func GetIDs(altIDs []*AltID) []string {
	ids := make([]string, 0, len(altIDs))
	for _, alt := range altIDs {
		ids = append(ids, alt.AltID)
	}
	return ids
}
