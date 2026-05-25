package heist

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/rbrabson/goblin/discord"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// A Theme is a set of messages that provide a "flavor" for a heist
type Theme struct {
	ID                  bson.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID             string          `json:"guild_id" bson:"guild_id"`
	Name                string          `json:"name" bson:"name"`
	EscapedMessages     []*HeistMessage `json:"escaped_messages" bson:"escaped_messages"`
	ApprehendedMessages []*HeistMessage `json:"apprehended_messages" bson:"apprehended_messages"`
	DiedMessages        []*HeistMessage `json:"died_messages" bson:"died_messages"`
	Jail                string          `json:"jail" bson:"jail"`
	OOB                 string          `json:"oob" bson:"oob"`
	Police              string          `json:"police" bson:"police"`
	Bail                string          `json:"bail" bson:"bail"`
	Crew                string          `json:"crew" bson:"crew"`
	Sentence            string          `json:"sentence" bson:"sentence"`
	Heist               string          `json:"heist" bson:"heist"`
	Vault               string          `json:"vault" bson:"vault"`
}

// A HeistMessage is a message for a successful heist outcome
type HeistMessage struct {
	Message     string       `json:"message" bson:"message"`
	BonusAmount int          `json:"bonus_amount,omitempty" bson:"bonus_amount,omitempty"`
	Result      MemberStatus `json:"result" bson:"result"`
}

// GetThemeNames returns a list of available themes
func GetThemeNames(guildID string) ([]string, error) {
	var names []string

	themes, err := readAllThemes(guildID)
	if err != nil {
		return nil, err
	}
	for _, theme := range themes {
		names = append(names, theme.Name)
	}

	return names, nil
}

// GetThemes returns all themes for a guild.
func GetThemes(guildID string) []*Theme {
	themes, err := readAllThemes(guildID)
	if err != nil {
		slog.Warn("unable to read themes",
			slog.String("guildID", guildID),
			slog.Any("error", err),
		)
		return nil
	}

	return themes
}

// GetTheme returns the theme for a guild
func GetTheme(guildID string, themeName string) *Theme {
	theme, err := readTheme(guildID, themeName)
	if err == nil && theme != nil {
		return theme
	}
	slog.Warn("unable to read theme",
		slog.String("guildID", guildID),
		slog.String("theme", themeName),
		slog.Any("error", err),
	)

	// The theme was not found in the DB, so create the default theme and use that
	theme = readThemeFromFile(guildID, themeName)
	if theme == nil {
		return nil
	}

	writeTheme(theme)
	slog.Debug("created default theme",
		slog.String("guildID", guildID),
		slog.String("theme", theme.Name),
	)

	return theme
}

// readThemeFromFile returns the default theme for a guild. If the theme can't be read
// from the configuration file or can't be decoded, nil is returned.
func readThemeFromFile(guildID string, themeName string) *Theme {
	configFileName := filepath.Join(discord.ConfigDir, "heist", "themes", themeName+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		slog.Error("failed to read default theme",
			slog.String("guildID", guildID),
			slog.String("file", configFileName),
			slog.Any("error", err),
		)
		return nil
	}

	theme := &Theme{}
	err = json.Unmarshal(bytes, theme)
	if err != nil {
		slog.Error("failed to unmarshal default theme",
			slog.String("guildID", guildID),
			slog.String("file", configFileName),
			slog.String("data", string(bytes)),
			slog.Any("error", err),
		)
		return nil
	}
	theme.GuildID = guildID
	theme.Name = HEIST_THEME

	slog.Debug("create new theme",
		slog.String("guildID", theme.GuildID),
		slog.String("theme", theme.Name),
	)

	return theme
}

// String returns a string representation of the Theme.
func (theme *Theme) String() string {
	if theme == nil {
		return "Theme<nil>"
	}

	return fmt.Sprintf("Theme{ID=%s, GuildID=%s, ThemeID=%s, Escaped=%d, Apprehended=%d, Died=%d, Jail=%s, OOB=%s, Police=%s, Bail=%s, Crew=%s, Sentence=%s, Heist=%s, Vault=%s}",
		theme.ID,
		theme.GuildID,
		theme.Name,
		len(theme.EscapedMessages),
		len(theme.ApprehendedMessages),
		len(theme.DiedMessages),
		theme.Jail,
		theme.OOB,
		theme.Police,
		theme.Bail,
		theme.Crew,
		theme.Sentence,
		theme.Heist,
		theme.Vault,
	)
}
