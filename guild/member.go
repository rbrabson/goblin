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

	return member
}

// SetName updates the name of the member as known on this guild (server).
func (member *Member) SetName(username string, nickname string, globalname string) *Member {
	slog.Debug("setting member name",
		slog.String("guildID", member.GuildID),
		slog.String("memberID", member.MemberID),
		slog.String("username", username),
		slog.String("nickname", nickname),
		slog.String("globalname", globalname),
	)

	var name string
	nickname = strings.Trim(nickname, " ")
	if strings.HasPrefix(nickname, "<") || strings.HasPrefix(nickname, "@") || strings.HasPrefix(nickname, "&") {
		slog.Debug("ignoring nickname",
			slog.String("nickname", nickname),
		)
		nickname = ""
	}
	globalname = strings.Trim(globalname, " ")
	if strings.HasPrefix(globalname, "<") || strings.HasPrefix(globalname, "@") || strings.HasPrefix(globalname, "&") {
		slog.Debug("ignoring globalname",
			slog.String("globalname", globalname),
		)
		globalname = ""
	}
	switch {
	case nickname != "":
		name = nickname
		slog.Debug("using nickname as name",
			slog.String("nickname", nickname),
		)
	case globalname != "":
		name = globalname
		slog.Debug("using globalname as name",
			slog.String("globalname", globalname),
		)
	default:
		name = username
		slog.Debug("using username as name",
			slog.String("username", username),
		)
	}
	if member.Name != name || member.UserName != username || member.NickName != nickname || member.GlobalName != globalname {
		member.Name = name
		member.UserName = username
		member.NickName = nickname
		member.GlobalName = globalname
		if err := writeMember(member); err != nil {
			slog.Error("failed to write member",
				slog.Any("error", err),
			)
		}
		slog.Debug("set member name",
			slog.String("guildID", member.GuildID),
			slog.String("memberID", member.MemberID),
			slog.String("name", member.Name),
		)
	} else {
		slog.Debug("member name unchanged",
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
	if err := writeMember(member); err != nil {
		slog.Error("failed to write member",
			slog.Any("error", err),
		)
	}
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
