package shop

import (
	"slices"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	SHOP_BAN = "shop"
)

// Member represents a member of a guild with restrictions on what they can or cannot do in a shop.
type Member struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID      string             `json:"guild_id,omitempty" bson:"guild_id,omitempty"`
	MemberID     string             `json:"member_id,omitempty" bson:"member_id,omitempty"`
	Restrictions []string           `json:"restrictions,omitempty" bson:"restrictions,omitempty"`
}

// GetMember retrieves a member from the database, creating one if it doesn't exist.
func GetMember(guildID, memberID string) *Member {
	member, err := readMember(guildID, memberID)
	if err != nil {
		member = newMember(guildID, memberID)
	}
	return member
}

// getMember retrieves a member from the database.
func getMember(guildID, memberID string) (*Member, error) {
	return readMember(guildID, memberID)
}

// NewMember creates a new member with the given guild ID and member ID.
func newMember(guildID, memberID string) *Member {
	return &Member{
		GuildID:      guildID,
		MemberID:     memberID,
		Restrictions: []string{},
	}
}

// AddRestriction adds a restriction to the member.
func (m *Member) AddRestriction(restriction string) error {
	m.Restrictions = append(m.Restrictions, restriction)
	return writeMember(m)
}

// RemoveRestriction removes a restriction from the member.
func (m *Member) RemoveRestriction(restriction string) error {
	for i, r := range m.Restrictions {
		if r == restriction {
			m.Restrictions = append(m.Restrictions[:i], m.Restrictions[i+1:]...)
			break
		}
	}

	if len(m.Restrictions) == 0 {
		return deleteMember(m)
	} else {
		return writeMember(m)
	}
}

// HasRestriction checks if the member has a specific restriction.
func (m *Member) HasRestriction(restriction string) bool {
	return slices.Contains(m.Restrictions, restriction)
}
