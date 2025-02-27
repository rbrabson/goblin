package race

import (
	"github.com/rbrabson/goblin/guild"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Racer represents a character that may be assigned to a member that partipates in a race
type Racer struct {
	ID            primitive.ObjectID `json:"_id" bson:"_id"`
	GuildID       string             `json:"guild_id" bson:"guild_id"`
	Theme         string             `json:"theme" bson:"theme"`
	Emoji         string             `json:"emoji" bson:"emoji"`
	MovementSpeed string             `json:"movement_speed" bson:"movement_speed"`
}

// GetRacers returns the list of chracters that may be assigned to a member during a race.
func GetRacers(g *guild.Guild, themeName string) []*Racer {
	log.Trace("--> race.GetRacers")
	defer log.Trace("<-- race.GetRacers")

	characters, err := getRacers(g, themeName)
	if err != nil {
		characters = newRacers(g)
	}
	return characters
}

// getRacers reads the list of characters for the theme and guild from the database. If the list
// does not exist, then an error is returned.
func getRacers(guild *guild.Guild, themeName string) ([]*Racer, error) {
	log.Trace("--> race.getRacers")
	defer log.Trace("<-- race.getRacers")

	filter := bson.D{{Key: "guild_id", Value: guild.GuildID}, {Key: "theme", Value: themeName}}
	racer, err := readAllRacers(filter)
	if err != nil {
		return nil, err
	}

	return racer, nil
}

// newRacers creates a new list of characters for the guild. The list is saved to
// the database.
func newRacers(g *guild.Guild) []*Racer {
	log.Trace("--> race.newRacers")
	defer log.Trace("<-- race.newRacers")

	racers := []*Racer{
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:minion:288380851023249408>",
			MovementSpeed: "veryfast",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:miner:288434873629147158>",
			MovementSpeed: "veryfast",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:goblin:288380850943295488>",
			MovementSpeed: "veryfast",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:betaminion:1138954584744808548>",
			MovementSpeed: "veryfast",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:wallbreaker:288380850620334080>",
			MovementSpeed: "fast",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:valkyrie:288380850738036746>",
			MovementSpeed: "fast",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:sneakyarcher:316157730434056193>",
			MovementSpeed: "fast",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:hogrider:1153765836604067980>",
			MovementSpeed: "fast",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:Queen:700065300137246790>",
			MovementSpeed: "fast",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:archer:1155943729882988654>",
			MovementSpeed: "fast",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:barbarian:288380850117148682>",
			MovementSpeed: "steady",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:cannoncart:693145555832012851>",
			MovementSpeed: "steady",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:healer:288380850662408203>",
			MovementSpeed: "steady",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:wizard:288380840289894401>",
			MovementSpeed: "steady",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:barbarianking:1138953623884267520>",
			MovementSpeed: "steady",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:GW:690611154061361183>",
			MovementSpeed: "steady",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:battlemachine:1155944842103369829>",
			MovementSpeed: "steady",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:bomber:1155945132735082539>",
			MovementSpeed: "abberant",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:dropship:316157731264397312>",
			MovementSpeed: "abberant",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:ballooncoc:288380851090096148>",
			MovementSpeed: "abberant",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:edragplead:872921108150616154>",
			MovementSpeed: "predator",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:dragoncoc:288380850402492416>",
			MovementSpeed: "predator",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:battleblimp:1153753935048347818>",
			MovementSpeed: "predator",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:lavahound:288380851090096138>",
			MovementSpeed: "predator",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:babydragon:342505061554978816>",
			MovementSpeed: "babydragon",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:ragedbarbarian:316157730735915009>",
			MovementSpeed: "special",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:superpekka:316157731302146050>",
			MovementSpeed: "slow",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:pekka:1153759228301946991>",
			MovementSpeed: "slow",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:bowler:288380850809339914>",
			MovementSpeed: "slow",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:witch:288380845830438923>",
			MovementSpeed: "slow",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:wallwrecker:935755961220616192>",
			MovementSpeed: "slow",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:nightwitch:316157731297820672>",
			MovementSpeed: "slow",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:golem:288380851232833546>",
			MovementSpeed: "slow",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:giant:288380850855477259>",
			MovementSpeed: "slow",
		},
		{
			GuildID:       g.GuildID,
			Theme:         "clash",
			Emoji:         "<:boxergiant:316157730782183426>",
			MovementSpeed: "slow",
		},
	}

	for _, racer := range racers {
		writeRacer(racer)
	}

	log.WithFields(log.Fields{"guild": g.GuildID, "count": len(racers)}).Info("created new racers")

	return racers
}
