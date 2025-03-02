package race

import (
	"math/rand"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Racer represents a character that may be assigned to a member that partipates in a race
type Racer struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID       string             `json:"guild_id" bson:"guild_id"`
	Theme         string             `json:"theme" bson:"theme"`
	Emoji         string             `json:"emoji" bson:"emoji"`
	MovementSpeed string             `json:"movement_speed" bson:"movement_speed"`
}

// GetRacers returns the list of chracters that may be assigned to a member during a race.
func GetRacers(guildID string, themeName string) []*Racer {
	log.Trace("--> race.GetRacers")
	defer log.Trace("<-- race.GetRacers")

	characters, err := getRacers(guildID, themeName)
	if err != nil {
		characters = newRacers(guildID)
	}
	return characters
}

// getRacers reads the list of characters for the theme and guild from the database. If the list
// does not exist, then an error is returned.
func getRacers(guildID string, themeName string) ([]*Racer, error) {
	log.Trace("--> race.getRacers")
	defer log.Trace("<-- race.getRacers")

	filter := bson.D{{Key: "guild_id", Value: guildID}, {Key: "theme", Value: themeName}}
	racer, err := readAllRacers(filter)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "theme": themeName, "error": err}).Error("unable to read racers")
		return nil, err
	}

	log.WithFields(log.Fields{"guild": guildID, "theme": themeName, "count": len(racer)}).Info("read racers")
	return racer, nil
}

// newRacers creates a new list of characters for the guild. The list is saved to
// the database.
func newRacers(guildID string) []*Racer {
	log.Trace("--> race.newRacers")
	defer log.Trace("<-- race.newRacers")

	racers := []*Racer{
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Minion:1345542060077486090>",
			MovementSpeed: "veryfast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Miner:1345543101388820481>",
			MovementSpeed: "veryfast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Goblin:1345542905388990464>",
			MovementSpeed: "veryfast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Beta_Minion:1345542031321075782>",
			MovementSpeed: "veryfast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Wall_Breaker:1345542069288177755>",
			MovementSpeed: "fast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Valkyrie:1345542914033455216>",
			MovementSpeed: "fast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Sneaky_Archer:1345543102756028476>",
			MovementSpeed: "fast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Hog_Rider:1345542907133952030>",
			MovementSpeed: "fast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Archer_Queen:1345546387819069481>",
			MovementSpeed: "fast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Archer:1345542023226327112>",
			MovementSpeed: "fast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Barbarian:1345542028146118767>",
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Cannon_Cart:1345542037751201843>",
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Healer:1345542051403530332>",
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Wizard:1345542076137214104>",
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Barbarian_King:1345542026506010645>",
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Grand_Warden:1345542906252886037>",
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Battle_Machine:1345542030188613662>",
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Bomber:1345542900691374090>",
			MovementSpeed: "abberant",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Dropship:1345542041442062448>",
			MovementSpeed: "abberant",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Balloon:B>",
			MovementSpeed: "abberant",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Electro_Dragon:1345542904315383838>",
			MovementSpeed: "predator",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Dragon:1345542903115550781>",
			MovementSpeed: "predator",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Battle_Blimp:1345542029194690600>",
			MovementSpeed: "predator",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Lava_Hound:1345542055501234206>",
			MovementSpeed: "predator",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Baby_Dragon:1345542024576630804>",
			MovementSpeed: "babydragon",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Raged_Barbarian:1345542062614905044>",
			MovementSpeed: "special",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Super_PEKKA:1345542065777541230>",
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Pekka:1345548786948374548>",
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Bowler:1345542034269798431>",
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Witch:1345542072769445908>",
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Wall_Wrecker:1345543103796477994>",
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Night_Witch:1345542910552313856>",
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Golem:1345542048090034290>",
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Giant:1345542044797636629>",
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         "<:Boxer_Giant:1345542901970505760>",
			MovementSpeed: "slow",
		},
	}

	/*
			racers := []*Racer{
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:minion:288380851023249408>",
				MovementSpeed: "veryfast",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:miner:288434873629147158>",
				MovementSpeed: "veryfast",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:goblin:288380850943295488>",
				MovementSpeed: "veryfast",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:betaminion:1138954584744808548>",
				MovementSpeed: "veryfast",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:wallbreaker:288380850620334080>",
				MovementSpeed: "fast",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:valkyrie:288380850738036746>",
				MovementSpeed: "fast",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:sneakyarcher:316157730434056193>",
				MovementSpeed: "fast",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:hogrider:1153765836604067980>",
				MovementSpeed: "fast",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:Queen:700065300137246790>",
				MovementSpeed: "fast",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:archer:1155943729882988654>",
				MovementSpeed: "fast",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:barbarian:288380850117148682>",
				MovementSpeed: "steady",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:cannoncart:693145555832012851>",
				MovementSpeed: "steady",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:healer:288380850662408203>",
				MovementSpeed: "steady",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:wizard:288380840289894401>",
				MovementSpeed: "steady",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:barbarianking:1138953623884267520>",
				MovementSpeed: "steady",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:GW:690611154061361183>",
				MovementSpeed: "steady",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:battlemachine:1155944842103369829>",
				MovementSpeed: "steady",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:bomber:1155945132735082539>",
				MovementSpeed: "abberant",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:dropship:316157731264397312>",
				MovementSpeed: "abberant",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:ballooncoc:288380851090096148>",
				MovementSpeed: "abberant",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:edragplead:872921108150616154>",
				MovementSpeed: "predator",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:dragoncoc:288380850402492416>",
				MovementSpeed: "predator",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:battleblimp:1153753935048347818>",
				MovementSpeed: "predator",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:lavahound:288380851090096138>",
				MovementSpeed: "predator",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:babydragon:342505061554978816>",
				MovementSpeed: "babydragon",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:ragedbarbarian:316157730735915009>",
				MovementSpeed: "special",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:superpekka:316157731302146050>",
				MovementSpeed: "slow",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:pekka:1153759228301946991>",
				MovementSpeed: "slow",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:bowler:288380850809339914>",
				MovementSpeed: "slow",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:witch:288380845830438923>",
				MovementSpeed: "slow",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:wallwrecker:935755961220616192>",
				MovementSpeed: "slow",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:nightwitch:316157731297820672>",
				MovementSpeed: "slow",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:golem:288380851232833546>",
				MovementSpeed: "slow",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:giant:288380850855477259>",
				MovementSpeed: "slow",
			},
			{
				GuildID:       guildID,
				Theme:         "clash",
				Emoji:         "<:boxergiant:316157730782183426>",
				MovementSpeed: "slow",
			},
		}
	*/

	for _, racer := range racers {
		writeRacer(racer)
	}

	log.WithFields(log.Fields{"guild": guildID, "count": len(racers)}).Info("created new racers")

	return racers
}

// calculateMovement calculates the distance a racer moves on a given turn
func (r *Racer) calculateMovement(currentTurn int) int {
	log.Trace("--> calculateMovement")
	defer log.Trace("<-- calculateMovement")

	switch r.MovementSpeed {
	case "veryfast":
		return rand.Intn(8) * 2
	case "fast":
		return rand.Intn(5) * 3
	case "slow":
		return (rand.Intn(3) + 1) * 3
	case "steady":
		return 2 * 3
	case "abberant":
		chance := rand.Intn(100)
		if chance > 90 {
			return 5 * 3
		}
		return rand.Intn(3) * 3
	case "predator":
		if currentTurn%2 == 0 {
			return 0
		} else {
			return (rand.Intn(4) + 2) * 3
		}
	case "special":
		fallthrough
	default:
		switch currentTurn {
		case 0:
			return 14 * 3
		case 1:
			return 0
		default:
			return rand.Intn(3) * 3
		}
	}
}
