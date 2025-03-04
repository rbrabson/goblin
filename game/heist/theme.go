package heist

import (
	"fmt"

	"github.com/rbrabson/goblin/internal/emoji"
	log "github.com/sirupsen/logrus"
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
	log.Trace("--> heist.GetThemes")
	defer log.Trace("<-- heist.GetThemes")

	themes, err := readAllThemes(guildID)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "error": err}).Error("unable to read themes")
		return nil
	}
	log.WithFields(log.Fields{"guild": guildID, "themes": len(themes)}).Trace("read all themes")
	return themes
}

// GetTheme returns the theme for a guild
func GetTheme(guildID string) *Theme {
	log.Trace("--> heist.GetTheme")
	defer log.Trace("<-- heist.GetTheme")

	config := GetConfig(guildID)
	theme, err := readTheme(guildID, config.Theme)
	if err == nil && theme != nil {
		log.WithFields(log.Fields{"guild": guildID, "theme": theme.Name}).Trace("read theme")
		return theme
	}
	log.WithFields(log.Fields{"guild": guildID, "error": err}).Error("unable to read theme")

	// The theme was found in the DB, so create the default theme and use that
	theme = getDefaultTheme(guildID)
	writeTheme(theme)
	log.WithFields(log.Fields{"guild": guildID, "theme": theme.Name}).Debug("created default theme")

	return theme
}

func getDefaultTheme(guildID string) *Theme {

	escapedMessages := []*HeistMessage{
		{
			Message:     "%s brought a few healers to keep themself alive " + emoji.Healer + ", +25 " + emoji.Gold,
			BonusAmount: 25,
			Result:      ESCAPED,
		},
		{
			Message:     "%s brought a rage spell to the raid " + emoji.RageSpell + ", +25 " + emoji.Gold,
			BonusAmount: 25,
			Result:      ESCAPED,
		},
		{
			Message:     "%s remembered to request CC troops before attacking. +25 " + emoji.Gold,
			BonusAmount: 25,
			Result:      ESCAPED,
		},
		{
			Message:     "%s used Royal Cloak " + emoji.ArcherQueen + ". +25 " + emoji.Gold,
			BonusAmount: 25,
			Result:      ESCAPED,
		},
		{
			Message:     "%s used Iron Fist " + emoji.BarbarianKing + ">. +25 " + emoji.Gold,
			BonusAmount: 25,
			Result:      ESCAPED,
		},
		{
			Message:     "%s took out a corner builder hut " + emoji.BuilderHut + ". +50 " + emoji.Gold + ">.",
			BonusAmount: 50,
			Result:      ESCAPED,
		},
		{
			Message:     "%s boosted the training barracks " + emoji.Gems + ". +50 " + emoji.Gold,
			BonusAmount: 50,
			Result:      ESCAPED,
		},
		{
			Message:     "%s lured defending CC troops. +50 " + emoji.Gold,
			BonusAmount: 50,
			Result:      ESCAPED,
		},
		{
			Message:     "%s built a funnel correctly. +100 " + emoji.Gold,
			BonusAmount: 100,
			Result:      ESCAPED,
		},
		{
			Message:     "%s used a power potion on the clan. +100 " + emoji.Gold,
			BonusAmount: 100,
			Result:      ESCAPED,
		},
		{
			Message:     "%s dropped a heal on a known Giant Bomb location " + emoji.HealingSpell + ". +100 " + emoji.Gold,
			BonusAmount: 100,
			Result:      ESCAPED,
		},
		{
			Message:     "%s successfully scouted the top base. +100 " + emoji.Gold,
			BonusAmount: 100,
			Result:      ESCAPED,
		},
		{
			Message:     "%s managed to take out the Town Hall. +150 " + emoji.Gold,
			BonusAmount: 150,
			Result:      ESCAPED,
		},
		{
			Message:     "%s got every last drop of Dark Elixir " + emoji.DarkElixer + ". +250 " + emoji.Gold,
			BonusAmount: 250,
			Result:      ESCAPED,
		},
		{
			Message:     "%s three starred a base with barch " + emoji.Barbarian + " " + emoji.SneakyArcher + ". +250 " + emoji.Gold,
			BonusAmount: 250,
			Result:      ESCAPED,
		},
		{
			Message:     "%s cast Eternal Tome " + emoji.GrandWarden + ". +250 " + emoji.Gold,
			BonusAmount: 250,
			Result:      ESCAPED,
		},
		{
			Message:     "%s HOG RIIIIIIIIDDEEEER " + emoji.HogRider + ". +500 " + emoji.Gold,
			BonusAmount: 500,
			Result:      ESCAPED,
		},
		{
			Message:     "%s scored a six pack " + emoji.SharkHi + ". +1,000 " + emoji.Gold,
			BonusAmount: 1000,
			Result:      ESCAPED,
		},
		{
			Message:     "%s 3 starred the top base in War without use of Heroes. +1,500 " + emoji.Gold,
			BonusAmount: 1500,
			Result:      ESCAPED,
		},
		{
			Message:     "%s found a gem box " + emoji.Gems + ". +2,500 " + emoji.Gold,
			BonusAmount: 2500,
			Result:      ESCAPED,
		},
	}

	apprehendedMessages := []*HeistMessage{
		{
			Message: "%s stepped onto a spring trap.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s forgot to bring heals to their GoHo attack " + emoji.Golem + " " + emoji.HogRider + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s used their builder potions while not upgrading anything.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s dropped rage on healers too late " + emoji.RageSpell + " " + emoji.Healer + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s forgot to bring heroes to the raid.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s forgot to bring CC troops to the raid.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s lost connection!",
			Result:  APPREHENDED,
		},
		{
			Message: "%s brought jump spells to a LaLo attack. " + emoji.JumpSpell + " " + emoji.LavaHound + " " + emoji.Balloon + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s fell out of a balloon " + emoji.Balloon + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s tried using a cold air balloon " + emoji.Balloon + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s got stuck in the clouds.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s stayed behind to finish taking out a wall.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s was paralyzed by a Hidden Tesla.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s dropped a freeze spell on themself " + emoji.FreezeSpell + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s attempted to teach the Hog Rider to ride Sheep instead " + emoji.HogRider + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s forgot to attack " + emoji.StarFishBarb + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s managed to get 0 stars " + emoji.StarEmpty + emoji.StarEmpty + emoji.StarEmpty + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s decided to upgrade all their barracks at the same time.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s decided to upgrade all their spell factories at the same time.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s decided to heal the grass " + emoji.HealingSpell + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s dropped their spells instead of troops and rage quit.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s decided to chase around a butterfly " + emoji.PEKKA + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s got lost in a Miner tunnel " + emoji.Miner + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s left Grand Warden in air mode " + emoji.GrandWarden + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s left Grand Warden in ground mode " + emoji.GrandWarden + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s spent all their gold on an empty Wall Wrecker " + emoji.WallWrecker + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s spent all their gold on a Battle Blimp with a hole in it " + emoji.BattleBlimp + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s was taken out by a Sneaky Archer " + emoji.SneakyArcher + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s was knocked into next week by a Boxer Giant " + emoji.BoxerGiant + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s was blown off the map by an Air Sweeper.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s used GoWiPe it was super ineffective " + emoji.Golem + " " + emoji.Wizard + " " + emoji.PEKKA + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s attempted to clone the Dark Elixir Storage " + emoji.CloneSpell + " " + emoji.DarkElixer + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s brought their farming army to the war " + emoji.Barbarian + " " + emoji.SneakyArcher + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s fell asleep under a builder hut " + emoji.Sleeping + " " + emoji.BuilderHut + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s got paralyzed by an Electro Dragon " + emoji.ElectroDragon + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s donated wallbreakers for defending CC " + emoji.WallBreaker + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s's dragons took a stroll around the perimeter of the base " + emoji.Dragon + " " + emoji.ElectroDragon + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s forgot to use the King's ability " + emoji.BarbarianKing + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s forgot to use the Queen's ability " + emoji.ArcherQueen + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s forgot to the use the Warden's ability " + emoji.GrandWarden + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s's wallbreakers went to the wrong wall " + emoji.WallBreaker + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s's night witches ran into a mega mine " + emoji.NightWitch + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s used a Book of Building to save the ten seconds needed to finish a level one gold mine.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s used a Book of Heroes to finish the level two Barbarian King upgrade " + emoji.BarbarianKing + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s spent gems on wall rings to get level two walls!",
			Result:  APPREHENDED,
		},
		{
			Message: "%s got a 49%% 0 Star. " + emoji.StarEmpty + emoji.StarEmpty + emoji.StarEmpty + ".",
			Result:  APPREHENDED,
		},
		{
			Message: "%s got a 99%% 1 Star in the deciding attack. " + emoji.StarNew + emoji.StarEmpty + emoji.StarEmpty + ".",
			Result:  APPREHENDED,
		},
	}
	diedMessages := []*HeistMessage{
		{
			Message: "%s forgot funnel and was singled out by an archer tower. (death) " + emoji.Toombstone,
			Result:  DEAD,
		},
		{
			Message: "%s walked into double giant bombs. (death) " + emoji.Toombstone,
			Result:  DEAD,
		},
		{
			Message: "%s was burnt to a crisp by a single target Inferno Tower. (death) " + emoji.Toombstone,
			Result:  DEAD,
		},
		{
			Message: "%s got killed by the CC troops. (death) " + emoji.Toombstone,
			Result:  DEAD,
		},
		{
			Message: "%s accidentally drank a poison spell " + emoji.PoisonSpell + ". (death) " + emoji.Toombstone,
			Result:  DEAD,
		},
		{
			Message: "%s tried to take out the Giga Tesla alone. (death) " + emoji.Toombstone,
			Result:  DEAD,
		},
		{
			Message: "%s was eaten by a dragon " + emoji.Dragon + ". (death) " + emoji.Toombstone,
			Result:  DEAD,
		},
		{
			Message: "%s wasn't healed by the healers " + emoji.Healer + ". (death) " + emoji.Toombstone,
			Result:  DEAD,
		},
		{
			Message: "%s pinged the discord mods for help with their attack. (death) " + emoji.Toombstone,
			Result:  DEAD,
		},
		{
			Message: "%s was tripled by an engineer. (death) " + emoji.Toombstone,
			Result:  DEAD,
		},
		{
			Message: "%s was banned by spAnser. (death) " + emoji.Toombstone,
			Result:  DEAD,
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
