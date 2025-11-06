package heist

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/internal/emoji"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// A Theme is a set of messages that provide a "flavor" for a heist
type Theme struct {
	ID                  primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID             string             `json:"guild_id" bson:"guild_id"`
	Name                string             `json:"name" bson:"name"`
	EscapedMessages     []*HeistMessage    `json:"escaped_messages" bson:"escaped_messages"`
	ApprehendedMessages []*HeistMessage    `json:"apprehended_messages" bson:"apprehended_messages"`
	DiedMessages        []*HeistMessage    `json:"died_messages" bson:"died_messages"`
	Jail                string             `json:"jail" bson:"jail"`
	OOB                 string             `json:"oob" bson:"oob"`
	Police              string             `json:"police" bson:"police"`
	Bail                string             `json:"bail" bson:"bail"`
	Crew                string             `json:"crew" bson:"crew"`
	Sentence            string             `json:"sentence" bson:"sentence"`
	Heist               string             `json:"heist" bson:"heist"`
	Vault               string             `json:"vault" bson:"vault"`
}

// A HeistMessage is a message for a successful heist outcome
type HeistMessage struct {
	Message     string       `json:"message" bson:"message"`
	BonusAmount int          `json:"bonus_amount,omitempty" bson:"bonus_amount,omitempty"`
	Result      MemberStatus `json:"result" bson:"result"`
}

// GetThemeNames returns a list of available themes
func GetThemeNames(guildID string) ([]string, error) {
	var fileNames []string
	themes := GetThemes(guildID)
	for _, theme := range themes {
		fileNames = append(fileNames, theme.Name)
	}

	return fileNames, nil
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
func GetTheme(guildID string) *Theme {
	config := GetConfig(guildID)
	theme, err := readTheme(guildID, config.Theme)
	if err == nil && theme != nil {
		return theme
	}
	slog.Error("unable to read theme",
		slog.String("guildID", guildID),
		slog.Any("error", err),
	)
	// The theme was not found in the DB, so create the default theme and use that

	// The theme was found in the DB, so create the default theme and use that
	theme = readThemeFromFile(guildID)
	writeTheme(theme)
	slog.Debug("created default theme",
		slog.String("guildID", guildID),
		slog.String("theme", theme.Name),
	)

	return theme
}

// readThemeFromFile returns the default theme for a guild. If the theme can't be read
// from the configuration file or can't be decoded, then a default theme is returned
func readThemeFromFile(guildID string) *Theme {
	configFileName := filepath.Join(discord.DISCORD_CONFIG_DIR, "heist", "themes", heistTheme+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		slog.Error("failed to read default theme",
			slog.String("guildID", guildID),
			slog.String("file", configFileName),
			slog.Any("error", err),
		)
		return getDefauiltTheme(guildID)
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
		return getDefauiltTheme(guildID)
	}
	theme.GuildID = guildID
	theme.Name = heistTheme

	slog.Info("create new theme",
		slog.String("guildID", theme.GuildID),
		slog.String("theme", theme.Name),
	)

	return theme
}

// getDefauiltTheme returns a default theme for the given guild.
func getDefauiltTheme(guildID string) *Theme {
	escapedMessages := []*HeistMessage{
		{
			Message:     "%s brought a few healers to keep themself alive " + emoji.Healer + ", +25 " + emoji.Gold,
			BonusAmount: 25,
			Result:      Escaped,
		},
		{
			Message:     "%s brought a rage spell to the raid " + emoji.RageSpell + ", +25 " + emoji.Gold,
			BonusAmount: 25,
			Result:      Escaped,
		},
		{
			Message:     "%s remembered to request CC troops before attacking. +25 " + emoji.Gold,
			BonusAmount: 25,
			Result:      Escaped,
		},
		{
			Message:     "%s used Royal Cloak " + emoji.ArcherQueen + ". +25 " + emoji.Gold,
			BonusAmount: 25,
			Result:      Escaped,
		},
		{
			Message:     "%s used Iron Fist " + emoji.BarbarianKing + ". +25 " + emoji.Gold,
			BonusAmount: 25,
			Result:      Escaped,
		},
		{
			Message:     "%s took out a corner builder hut " + emoji.BuilderHut + ". +50 " + emoji.Gold + ".",
			BonusAmount: 50,
			Result:      Escaped,
		},
		{
			Message:     "%s boosted the training barracks " + emoji.Gems + ". +50 " + emoji.Gold,
			BonusAmount: 50,
			Result:      Escaped,
		},
		{
			Message:     "%s lured defending CC troops. +50 " + emoji.Gold,
			BonusAmount: 50,
			Result:      Escaped,
		},
		{
			Message:     "%s built a funnel correctly. +100 " + emoji.Gold,
			BonusAmount: 100,
			Result:      Escaped,
		},
		{
			Message:     "%s used a power potion on the clan. +100 " + emoji.Gold,
			BonusAmount: 100,
			Result:      Escaped,
		},
		{
			Message:     "%s dropped a heal on a known Giant Bomb location " + emoji.HealingSpell + ". +100 " + emoji.Gold,
			BonusAmount: 100,
			Result:      Escaped,
		},
		{
			Message:     "%s successfully scouted the top base. +100 " + emoji.Gold,
			BonusAmount: 100,
			Result:      Escaped,
		},
		{
			Message:     "%s managed to take out the Town Hall. +150 " + emoji.Gold,
			BonusAmount: 150,
			Result:      Escaped,
		},
		{
			Message:     "%s got every last drop of Dark Elixir " + emoji.DarkElixer + ". +250 " + emoji.Gold,
			BonusAmount: 250,
			Result:      Escaped,
		},
		{
			Message:     "%s three starred a base with barch " + emoji.Barbarian + " " + emoji.SneakyArcher + ". +250 " + emoji.Gold,
			BonusAmount: 250,
			Result:      Escaped,
		},
		{
			Message:     "%s cast Eternal Tome " + emoji.GrandWarden + ". +250 " + emoji.Gold,
			BonusAmount: 250,
			Result:      Escaped,
		},
		{
			Message:     "%s HOG RIIIIIIIIDDEEEER " + emoji.HogRider + ". +500 " + emoji.Gold,
			BonusAmount: 500,
			Result:      Escaped,
		},
		{
			Message:     "%s scored a six pack " + emoji.SharkHi + ". +1,000 " + emoji.Gold,
			BonusAmount: 1000,
			Result:      Escaped,
		},
		{
			Message:     "%s 3 starred the top base in War without use of Heroes. +1,500 " + emoji.Gold,
			BonusAmount: 1500,
			Result:      Escaped,
		},
		{
			Message:     "%s found a gem box " + emoji.Gems + ". +2,500 " + emoji.Gold,
			BonusAmount: 2500,
			Result:      Escaped,
		},
	}

	apprehendedMessages := []*HeistMessage{
		{
			Message: "%s stepped onto a spring trap.",
			Result:  Apprehended,
		},
		{
			Message: "%s forgot to bring heals to their GoHo attack " + emoji.Golem + " " + emoji.HogRider + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s used their builder potions while not upgrading anything.",
			Result:  Apprehended,
		},
		{
			Message: "%s dropped rage on healers too late " + emoji.RageSpell + " " + emoji.Healer + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s forgot to bring heroes to the raid.",
			Result:  Apprehended,
		},
		{
			Message: "%s forgot to bring CC troops to the raid.",
			Result:  Apprehended,
		},
		{
			Message: "%s lost connection!",
			Result:  Apprehended,
		},
		{
			Message: "%s brought jump spells to a LaLo attack. " + emoji.JumpSpell + " " + emoji.LavaHound + " " + emoji.Balloon + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s fell out of a balloon " + emoji.Balloon + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s tried using a cold air balloon " + emoji.Balloon + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s got stuck in the clouds.",
			Result:  Apprehended,
		},
		{
			Message: "%s stayed behind to finish taking out a wall.",
			Result:  Apprehended,
		},
		{
			Message: "%s was paralyzed by a Hidden Tesla.",
			Result:  Apprehended,
		},
		{
			Message: "%s dropped a freeze spell on themself " + emoji.FreezeSpell + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s attempted to teach the Hog Rider to ride Sheep instead " + emoji.HogRider + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s forgot to attack " + emoji.StarFishBarb + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s managed to get 0 stars " + emoji.StarEmpty + emoji.StarEmpty + emoji.StarEmpty + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s decided to upgrade all their barracks at the same time.",
			Result:  Apprehended,
		},
		{
			Message: "%s decided to upgrade all their spell factories at the same time.",
			Result:  Apprehended,
		},
		{
			Message: "%s decided to heal the grass " + emoji.HealingSpell + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s dropped their spells instead of troops and rage quit.",
			Result:  Apprehended,
		},
		{
			Message: "%s decided to chase around a butterfly " + emoji.PEKKA + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s got lost in a Miner tunnel " + emoji.Miner + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s left GW in air mode " + emoji.GrandWarden + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s left GW in ground mode " + emoji.GrandWarden + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s spent all their gold on an empty Wall Wrecker " + emoji.WallWrecker + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s spent all their gold on a Battle Blimp with a hole in it " + emoji.BattleBlimp + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s was taken out by a Sneaky Archer " + emoji.SneakyArcher + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s was knocked into next week by a Boxer Giant " + emoji.BoxerGiant + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s was blown off the map by an Air Sweeper.",
			Result:  Apprehended,
		},
		{
			Message: "%s used GoWiPe it was super ineffective " + emoji.Golem + " " + emoji.Wizard + " " + emoji.PEKKA + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s attempted to clone the Dark Elixir Storage " + emoji.CloneSpell + " " + emoji.DarkElixer + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s brought their farming army to the war " + emoji.Barbarian + " " + emoji.SneakyArcher + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s fell asleep under a builder hut " + emoji.Sleeping + " " + emoji.BuilderHut + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s got paralyzed by an Electro Dragon " + emoji.ElectroDragon + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s donated wallbreakers for defending CC " + emoji.WallBreaker + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s's dragons took a stroll around the perimeter of the base " + emoji.Dragon + " " + emoji.ElectroDragon + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s forgot to use the King's ability " + emoji.BarbarianKing + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s forgot to use the Queen's ability " + emoji.ArcherQueen + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s forgot to the use the Warden's ability " + emoji.GrandWarden + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s's wallbreakers went to the wrong wall " + emoji.WallBreaker + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s's night witches ran into a mega mine " + emoji.NightWitch + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s used a Book of Building to save the ten seconds needed to finish a level one gold mine.",
			Result:  Apprehended,
		},
		{
			Message: "%s used a Book of Heroes to finish the level two BK upgrade " + emoji.BarbarianKing + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s spent gems on wall rings to get level two walls!",
			Result:  Apprehended,
		},
		{
			Message: "%s got a 49%% 0 Star. " + emoji.StarEmpty + emoji.StarEmpty + emoji.StarEmpty + ".",
			Result:  Apprehended,
		},
		{
			Message: "%s got a 99%% 1 Star in the deciding attack. " + emoji.StarNew + emoji.StarEmpty + emoji.StarEmpty + ".",
			Result:  Apprehended,
		},
	}
	diedMessages := []*HeistMessage{
		{
			Message: "%s forgot funnel and was singled out by an archer tower. (death) " + emoji.Toombstone,
			Result:  Dead,
		},
		{
			Message: "%s walked into double giant bombs. (death) " + emoji.Toombstone,
			Result:  Dead,
		},
		{
			Message: "%s was burnt to a crisp by a single target Inferno Tower. (death) " + emoji.Toombstone,
			Result:  Dead,
		},
		{
			Message: "%s got killed by the CC troops. (death) " + emoji.Toombstone,
			Result:  Dead,
		},
		{
			Message: "%s accidentally drank a poison spell " + emoji.PoisonSpell + ". (death) " + emoji.Toombstone,
			Result:  Dead,
		},
		{
			Message: "%s tried to take out the Giga Tesla alone. (death) " + emoji.Toombstone,
			Result:  Dead,
		},
		{
			Message: "%s was eaten by a dragon " + emoji.Dragon + ". (death) " + emoji.Toombstone,
			Result:  Dead,
		},
		{
			Message: "%s wasn't healed by the healers " + emoji.Healer + ". (death) " + emoji.Toombstone,
			Result:  Dead,
		},
		{
			Message: "%s pinged the discord mods for help with their attack. (death) " + emoji.Toombstone,
			Result:  Dead,
		},
		{
			Message: "%s was tripled by an engineer. (death) " + emoji.Toombstone,
			Result:  Dead,
		},
		{
			Message: "%s was banned by spAnser. (death) " + emoji.Toombstone,
			Result:  Dead,
		},
	}

	return &Theme{
		GuildID:             guildID,
		Name:                "clash",
		EscapedMessages:     escapedMessages,
		ApprehendedMessages: apprehendedMessages,
		DiedMessages:        diedMessages,
		Jail:                "resting",
		OOB:                 "running on gems",
		Police:              "Enemy CC Troops",
		Bail:                "heal",
		Crew:                "clan",
		Sentence:            "nap",
		Heist:               "raid",
		Vault:               "village",
	}
}

// String returns a string representation of the Theme.
func (theme *Theme) String() string {
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
