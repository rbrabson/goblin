package heist

import (
	"github.com/rbrabson/dgame/guild"
	log "github.com/sirupsen/logrus"
)

const (
	THEME = "theme"
)

var (
	themes = make(map[string]map[string]*Theme)
)

// A Theme is a set of messages that provide a "flavor" for a heist
type Theme struct {
	ID       string        `json:"_id" bson:"_id"`
	Good     []GoodMessage `json:"good"`
	Bad      []BadMessage  `json:"bad"`
	Jail     string        `json:"jail" bson:"jail"`
	OOB      string        `json:"oob" bson:"oob"`
	Police   string        `json:"police" bson:"police"`
	Bail     string        `json:"bail" bson:"bail"`
	Crew     string        `json:"crew" bson:"crew"`
	Sentence string        `json:"sentence" bson:"sentence"`
	Heist    string        `json:"heist" bson:"heist"`
	Vault    string        `json:"vault" bson:"vault"`
	guildID  string        `json:"-" bson:"-"`
}

// A GoodMessage is a message for a successful heist outcome
type GoodMessage struct {
	Message string `json:"message" bson:"message"`
	Amount  int    `json:"amount" bson:"amount"`
}

// A BadMessage is a message for a failed heist outcome
type BadMessage struct {
	Message string       `json:"message" bson:"message"`
	Result  MemberStatus `json:"result" bson:"result"`
}

// LoadThemes loads all available themes for a guild
func LoadThemes(guild *guild.Guild) map[string]*Theme {
	log.Trace("--> heist.LoadThemes")
	defer log.Trace("<-- heist.LoadThemes")

	guildThemes := make(map[string]*Theme)
	themes[guild.ID] = guildThemes

	themeNames, _ := db.ListDocuments(guild.ID, THEME)
	for _, themeName := range themeNames {
		var theme Theme
		db.Read(guild.ID, THEME, themeName, &theme)
		theme.guildID = guild.ID
		guildThemes[theme.ID] = &theme
		log.WithFields(log.Fields{"guild": theme.guildID, "theme": theme.ID}).Debug("load theme from database")
	}

	return guildThemes
}

// GetTheme returns the named theme for the given guild
func GetTheme(guild *guild.Guild, themeName string) *Theme {
	log.Trace("--> heist.GetTheme")
	defer log.Trace("<-- heist.GetTheme")

	guildThemes := themes[guild.ID]
	if guildThemes == nil {
		guildThemes = make(map[string]*Theme)
		themes[guild.ID] = guildThemes
		log.WithFields(log.Fields{"guild": guild.ID, "theme": themeName}).Debug("create theme mapping")
	}

	theme := guildThemes[themeName]
	if theme == nil {
		log.WithFields(log.Fields{"guild": guild.ID, "theme": themeName}).Warn("theme not found")
	}

	return theme
}

// GetThemeNames returns a list of available themes
func GetThemeNames(guild *guild.Guild) ([]string, error) {
	var fileNames []string
	for _, theme := range themes[guild.ID] {
		fileNames = append(fileNames, theme.ID)
	}

	return fileNames, nil
}

// Write creates or updates the theme in the database
func (theme *Theme) Write() {
	log.Trace("--> heist.Theme.Write")
	defer log.Trace("<-- heist.Theme.Write")

	db.Write(theme.guildID, THEME, theme.ID, theme)
}
