package blackjack

import (
	"strings"

	bj "github.com/rbrabson/blackjack"
)

var (
	blackjackTables = make(map[string]*Table)
)

// Dealer represents the dealer in a blackjack game.
type Dealer struct {
	Hand *bj.Hand
}

// String returns a string representation of the Dealer.
func (d *Dealer) String() string {
	var sb strings.Builder
	sb.WriteString("Dealer{")
	if d.Hand != nil {
		sb.WriteString(d.Hand.String())
	} else {
		sb.WriteString("<nil>")
	}
	sb.WriteString("}")
	return sb.String()
}

// Player represents a player in a blackjack game.
type Player struct {
	Member *Member
	Hand   *bj.Hand
}

// String returns a string representation of the Player.
func (p *Player) String() string {
	var sb strings.Builder
	sb.WriteString("Player{")
	if p.Member != nil {
		sb.WriteString(p.Member.String())
	} else {
		sb.WriteString("<nil>")
	}
	sb.WriteString(", ")
	if p.Hand != nil {
		sb.WriteString(p.Hand.String())
	} else {
		sb.WriteString("<nil>")
	}
	sb.WriteString("}")
	return sb.String()
}

// Table represents a blackjack game.
type Table struct {
	GuildID string
	Dealer  *Dealer
	Players []*Player
	Config  *Config
}

// String returns a string representation of the blackjack Table.
func (g *Table) String() string {
	var sb strings.Builder
	sb.WriteString("Table{")
	sb.WriteString("Dealer: ")
	if g.Dealer != nil {
		sb.WriteString(g.Dealer.String())
	} else {
		sb.WriteString("<nil>")
	}
	sb.WriteString(", ")
	sb.WriteString("Players: [")
	for i, player := range g.Players {
		if i > 0 {
			sb.WriteString(", ")
		}
		if player != nil {
			sb.WriteString(player.String())
		} else {
			sb.WriteString("<nil>")
		}
	}
	sb.WriteString("], ")
	sb.WriteString("Config: ")
	sb.WriteString(g.Config.String())
	sb.WriteString("}")
	return sb.String()
}

// GetTable retrieves an existing blackjack table for the given guild ID or creates a new one if none exists.
func GetTable(guildID string) (*Table, error) {
	// TODO: check to see if there is a cooldown
	var table *Table
	if _, exists := blackjackTables[guildID]; !exists {
		table = newTable(guildID)
	} else {
		table = blackjackTables[guildID]
	}
	return table, nil
}

// newTable creates a new blackjack table for the given guild ID.
func newTable(guildID string) *Table {
	table := &Table{
		GuildID: guildID,
		Players: []*Player{},
		Dealer:  &Dealer{Hand: &bj.Hand{}},
		Config:  GetConfig(),
	}
	blackjackTables[guildID] = table
	return table
}
