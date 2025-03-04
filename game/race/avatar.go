package race

import (
	"math/rand"

	"github.com/rbrabson/goblin/internal/emoji"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RaceAvatar represents a character that may be assigned to a member that partipates in a race
type RaceAvatar struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID       string             `json:"guild_id" bson:"guild_id"`
	Theme         string             `json:"theme" bson:"theme"`
	Emoji         string             `json:"emoji" bson:"emoji"`
	MovementSpeed string             `json:"movement_speed" bson:"movement_speed"`
}

// GetRaceAvatars returns the list of chracters that may be assigned to a member during a race.
func GetRaceAvatars(guildID string, themeName string) []*RaceAvatar {
	log.Trace("--> race.GetRaceAvatars")
	defer log.Trace("<-- race.GetRaceAvatars")

	characters, err := getRaceAvatars(guildID, themeName)
	if err != nil {
		characters = newRaceAvatars(guildID)
	}
	return characters
}

// getRaceAvatars reads the list of characters for the theme and guild from the database. If the list
// does not exist, then an error is returned.
func getRaceAvatars(guildID string, themeName string) ([]*RaceAvatar, error) {
	log.Trace("--> race.getRaceAvatars")
	defer log.Trace("<-- race.getRaceAvatars")

	filter := bson.D{{Key: "guild_id", Value: guildID}, {Key: "theme", Value: themeName}}
	racer, err := readAllRacers(filter)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "theme": themeName, "error": err}).Error("unable to read racers")
		return nil, err
	}

	log.WithFields(log.Fields{"guild": guildID, "theme": themeName, "count": len(racer)}).Info("read racers")
	return racer, nil
}

// newRaceAvatars creates a new list of characters for the guild. The list is saved to
// the database.
func newRaceAvatars(guildID string) []*RaceAvatar {
	log.Trace("--> race.newRaceAvatars")
	defer log.Trace("<-- race.newRaceAvatars")

	racers := []*RaceAvatar{
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Minion,
			MovementSpeed: "veryfast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Miner,
			MovementSpeed: "veryfast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Goblin,
			MovementSpeed: "veryfast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.BetaMinion,
			MovementSpeed: "veryfast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.WallBreaker,
			MovementSpeed: "fast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Valkrie,
			MovementSpeed: "fast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.SneakyArcher,
			MovementSpeed: "fast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.HogRider,
			MovementSpeed: "fast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.ArcherQueen,
			MovementSpeed: "fast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Archer,
			MovementSpeed: "fast",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Barbarian,
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.CannonCart,
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Healer,
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Wizard,
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.BarbarianKing,
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.GrandWarden,
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.BattleMachine,
			MovementSpeed: "steady",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Bomber,
			MovementSpeed: "abberant",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.DropShip,
			MovementSpeed: "abberant",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Balloon,
			MovementSpeed: "abberant",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.ElectroDragon,
			MovementSpeed: "predator",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Dragon,
			MovementSpeed: "predator",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.BattleBlimp,
			MovementSpeed: "predator",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.LavaHound,
			MovementSpeed: "predator",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.BabyDragon,
			MovementSpeed: "babydragon",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.RagedBarbarian,
			MovementSpeed: "special",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.SuperPEKKA,
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.PEKKA,
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Bowler,
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Witch,
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.WallWrecker,
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.NightWitch,
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Golem,
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.Giant,
			MovementSpeed: "slow",
		},
		{
			GuildID:       guildID,
			Theme:         "clash",
			Emoji:         emoji.BoxerGiant,
			MovementSpeed: "slow",
		},
	}

	/*
		racers := []*RaceAvatar{
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
func (r *RaceAvatar) calculateMovement(currentTurn int) int {
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

// String returns a string representation of the race avatar.
func (r *RaceAvatar) String() string {
	return r.Emoji
}
