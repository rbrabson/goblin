package heist

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// A Theme is a set of messages that provide a "flavor" for a heist
type Theme struct {
	ID                  primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID             string             `json:"guild_id" bson:"guild_id"`
	Name                string             `json:"name" bson:"name"`
	EscapedMessages     []*GoodMessage     `json:"escaped_messages" bson:"escaped_messages"`
	ApprehendedMessages []*BadMessage      `json:"apprehended_messages" bson:"apprehended_messages"`
	DiedMessages        []*BadMessage      `json:"died_messages" bson:"died_messages"`
	Jail                string             `json:"jail" bson:"jail"`
	OOB                 string             `json:"oob" bson:"oob"`
	Police              string             `json:"police" bson:"police"`
	Bail                string             `json:"bail" bson:"bail"`
	Crew                string             `json:"crew" bson:"crew"`
	Sentence            string             `json:"sentence" bson:"sentence"`
	Heist               string             `json:"heist" bson:"heist"`
	Vault               string             `json:"vault" bson:"vault"`
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

	escapedMessages := []*GoodMessage{
		{
			Message: "%s brought a few healers to keep themself alive <:healer:288380850662408203>. +25 <:gold:312346438157991938>",
			Amount:  25,
		},
		{
			Message: "%s brought a rage spell to the raid <:ragespell:1153756488117014578>. +25 <:gold:312346438157991938>",
			Amount:  25,
		},
		{
			Message: "%s remembered to request CC troops before attacking. +25 <:gold:312346438157991938>",
			Amount:  25,
		},
		{
			Message: "%s used Royal Cloak <:Queen:700065300137246790>. +25 <:gold:312346438157991938>",
			Amount:  25,
		},
		{
			Message: "%s used Iron Fist <:BK:288380851111329792>. +25 <:gold:312346438157991938>",
			Amount:  25,
		},
		{
			Message: "%s took out a corner builder hut <:builderhut:1153757245914488914>. +50 <:gold:312346438157991938>.",
			Amount:  50,
		},
		{
			Message: "%s boosted the training barracks <:gems:312346463453708289>. +50 <:gold:312346438157991938>",
			Amount:  50,
		},
		{
			Message: "%s lured defending CC troops. +50 <:gold:312346438157991938>",
			Amount:  50,
		},
		{
			Message: "%s built a funnel correctly. +100 <:gold:312346438157991938>",
			Amount:  100,
		},
		{
			Message: "%s used a power potion on the clan. +100 <:gold:312346438157991938>",
			Amount:  100,
		},
		{
			Message: "%s dropped a heal on a known Giant Bomb location <:healingspell:1153756342016823316>. +100 <:gold:312346438157991938>",
			Amount:  100,
		},
		{
			Message: "%s successfully scouted the top base. +100 <:gold:312346438157991938>",
			Amount:  100,
		},
		{
			Message: "%s managed to take out the Town Hall. +150 <:gold:312346438157991938>",
			Amount:  150,
		},
		{
			Message: "%s got every last drop of Dark Elixir <:darkelixir:312346454645669889>. +250 <:gold:312346438157991938>",
			Amount:  250,
		},
		{
			Message: "%s three starred a base with barch <:barbarian:288380850117148682> <:sneakyarcher:316157730434056193>. +250 <:gold:312346438157991938>",
			Amount:  250,
		},
		{
			Message: "%s cast Eternal Tome <:GW:690611154061361183>. +250 <:gold:312346438157991938>",
			Amount:  250,
		},
		{
			Message: "%s HOG RIIIIIIIIDDEEEER <:hogrider:1153765836604067980>. +500 <:gold:312346438157991938>",
			Amount:  500,
		},
		{
			Message: "%s scored a six pack <:sharkhi:301544708063100930>. +1,000 <:gold:312346438157991938>",
			Amount:  1000,
		},
		{
			Message: "%s 3 starred the top base in War without use of Heroes. +1,500 <:gold:312346438157991938>",
			Amount:  1500,
		},
		{
			Message: "%s found a gem box <:gems:312346463453708289>. +2,500 <:gold:312346438157991938>",
			Amount:  2500,
		},
	}

	apprehendedMessages := []*BadMessage{
		{
			Message: "%s stepped onto a spring trap.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s forgot to bring heals to their GoHo attack <:golem:288380851232833546><:hogrider:1153765836604067980>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s used their builder potions while not upgrading anything.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s dropped rage on healers too late <:ragespell:1153756488117014578> <:healer:288380850662408203>.",
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
			Message: "%s brought jump spells to a LaLo attack. <:jumpspell:1153756390934978613> <:lavahound:288380851090096138> <:ballooncoc:288380851090096148> ",
			Result:  APPREHENDED,
		},
		{
			Message: "%s fell out of a balloon <:ballooncoc:288380851090096148>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s tried using a cold air balloon <:ballooncoc:288380851090096148>.",
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
			Message: "%s dropped a freeze spell on themself <:freezespell:1153757695078301797>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s attempted to teach the Hog Rider to ride Sheep instead <:hogrider:1153765836604067980>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s forgot to attack <:starfishbarb:312350355960889344>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s managed to get 0 stars.",
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
			Message: "%s decided to heal the grass <:healingspell:1153756342016823316>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s dropped their spells instead of troops and rage quit.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s decided to chase around a butterfly <:pekka:1153759228301946991>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s got lost in a Miner tunnel <:miner:288434873629147158>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s left Grand Warden in air mode <:GW:690611154061361183>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s left Grand Warden in ground mode <:GW:690611154061361183>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s spent all their gold on an empty Wall Wrecker <:wallwrecker:935755961220616192>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s spent all their gold on a Battle Blimp with a hole in it <:battleblimp:1153753935048347818>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s was taken out by a Sneaky Archer <:sneakyarcher:316157730434056193>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s was knocked into next week by a Boxer Giant <:boxergiant:316157730782183426>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s was blown off the map by an Air Sweeper.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s used GoWiPe it was super ineffective <:golem:288380851232833546><:wizard:288380840289894401><:pekka:1153759228301946991>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s attempted to clone the Dark Elixir Storage <:clonespell:1153754704401141760> <:darkelixir:312346454645669889>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s brought their farming army to the war <:barbarian:288380850117148682> <:sneakyarcher:316157730434056193>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s fell asleep under a builder hut <a:sleep_zzz:400680923151990784>  <:builderhut:1153757245914488914>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s got paralyzed by an Electro Dragon <:edragBplease:721984769667366919>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s donated wallbreakers for defending CC <:wallwrecker:935755961220616192>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s's dragons took a stroll around the perimeter of the base <:dragoncoc:288380850402492416> <:edragBplease:721984769667366919>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s forgot to use the King's ability <:BK:288380851111329792>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s forgot to use the Queen's ability <:Queen:700065300137246790>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s forgot to the use the Warden's ability <:GW:690611154061361183>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s's wallbreakers went to the wrong wall <:wallwrecker:935755961220616192>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s's night witches ran into a mega mine <:nightwitch:316157731297820672>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s used a Book of Building to save the ten seconds needed to finish a level one gold mine.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s used a Book of Heroes to finish the level two Barbarian King upgrade <:BK:288380851111329792>.",
			Result:  APPREHENDED,
		},
		{
			Message: "%s spent gems on wall rings to get level two walls!",
			Result:  APPREHENDED,
		},
		{
			Message: "%s got a 49%% 0 Star. <:starempty:1153758399117414502><:starempty:1153758399117414502><:starempty:1153758399117414502>",
			Result:  APPREHENDED,
		},
		{
			Message: "%s got a 99%% 1 Star in the deciding attack. <:starnew:1153758461553807480><:starempty:1153758399117414502><:starempty:1153758399117414502>",
			Result:  APPREHENDED,
		},
	}
	diedMessages := []*BadMessage{
		{
			Message: "%s forgot funnel and was singled out by an archer tower. (death) <:tombstone:688467782496419942>",
			Result:  DEAD,
		},
		{
			Message: "%s walked into double giant bombs. (death) <:tombstone:688467782496419942>",
			Result:  DEAD,
		},
		{
			Message: "%s was burnt to a crisp by a single target Inferno Tower. (death) <:tombstone:688467782496419942>",
			Result:  DEAD,
		},
		{
			Message: "%s got killed by the CC troops. (death) <:tombstone:688467782496419942>",
			Result:  DEAD,
		},
		{
			Message: "%s accidentally drank a poison spell <:poisonspell:1153756439496634389>. (death) <:tombstone:688467782496419942>",
			Result:  DEAD,
		},
		{
			Message: "%s tried to take out the Giga Tesla alone. (death) <:tombstone:688467782496419942>",
			Result:  DEAD,
		},
		{
			Message: "%s was eaten by a dragon <:dragoncoc:288380850402492416>. (death) <:tombstone:688467782496419942>",
			Result:  DEAD,
		},
		{
			Message: "%s wasn't healed by the healers <:healer:288380850662408203>. (death) <:tombstone:688467782496419942>",
			Result:  DEAD,
		},
		{
			Message: "%s pinged the discord mods for help with their attack. (death) <:tombstone:688467782496419942>",
			Result:  DEAD,
		},
		{
			Message: "%s was tripled by an engineer. (death) <:tombstone:688467782496419942>",
			Result:  DEAD,
		},
		{
			Message: "%s was banned by spAnser. (death) <:tombstone:688467782496419942>",
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
