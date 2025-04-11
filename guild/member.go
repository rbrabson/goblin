package guild

import (
	"fmt"
	"log/slog"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Member is a member of a given guild
type Member struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID    string             `json:"guild_id" bson:"guild_id"`
	MemberID   string             `json:"member_id" bson:"member_id"`
	UserName   string             `json:"username" bson:"username"`
	GlobalName string             `json:"global_name" bson:"global_name"`
	NickName   string             `json:"nickname" bson:"nickname"`
	Name       string             `json:"name" bson:"name"`
}

// GetMember returns a member in the guild (server). If one doesnt' exist, then one is created with a blank name.
func GetMember(guildID string, memberID string) *Member {
	member := readMember(guildID, memberID)

	if member == nil {
		member = newMember(guildID, memberID)
	}
	// if member.UserName == "" || member.Name == "" {
	// 	guildMember, err := s.GuildMember(guildID, memberID)
	// 	if err != nil {
	// 		slog.WithFields(slog.Fields{"guild": guildID, "member": memberID, "error": err}).Error("failed to get guild member")
	// 		return member
	// 	}
	// 	member.SetName(guildMember.User.Username, guildMember.Nick, guildMember.User.GlobalName)
	// 	slog.WithFields(slog.Fields{"guild": guildID, "member": memberID, "nickname": guildMember.Nick, "username": guildMember.User.Username, "userid": guildMember.User.ID, "globalname": guildMember.User.GlobalName}).Debug("updated member")
	// }

	return member
}

// SetName updates the name of the member as known on this guild (server).
func (member *Member) SetName(username string, nickname string, globalname string) *Member {
	var name string
	switch {
	case nickname != "":
		name = nickname
	case globalname != "":
		name = globalname
	default:
		name = username
	}
	if member.Name != name || member.UserName != username || member.NickName != nickname || member.GlobalName != globalname {
		member.Name = strings.Trim(name, "# ")
		member.UserName = username
		member.NickName = nickname
		member.GlobalName = globalname
		writeMember(member)
		slog.Debug("set member name",
			slog.String("guildID", member.GuildID),
			slog.String("memberID", member.MemberID),
			slog.String("name", member.Name),
		)
	}

	return member
}

// newMember creates a new member in the guild (server).
func newMember(guildID string, memberID string) *Member {
	member := &Member{
		MemberID: memberID,
		GuildID:  guildID,
	}
	writeMember(member)
	slog.Info("created new member",
		slog.String("guildID", member.GuildID),
		slog.String("memberID", member.MemberID),
	)

	return member
}

// String returns a string representation of the Member.
func (member *Member) String() string {
	return fmt.Sprintf("Member{ID=%s, GuildID=%s, MemberID=%s, Name=%s}",
		member.ID.Hex(),
		member.GuildID,
		member.MemberID,
		member.Name)
}
