package blackjack

// Table represents a blackjack table. Players can join the table and play blackjack.
type Table struct {
	Shoe    *Shoe
	Dealer  *Hand
	Players []*Player
	Config  *Config
}

// NewTable returns a new table with a shoe and dealer.
func NewTable(guildID string, numDecks int) *Table {
	config := GetConfig(guildID)
	return &Table{
		Shoe:    NewShoe(numDecks),
		Players: make([]*Player, 0),
		Config:  config,
	}
}

// AddPlayer adds a player to the table.
func (table *Table) AddPlayer(player *Player) {
	table.Players = append(table.Players, player)
	player.Table = table
}

// RemovePlayer removes a player from the table.
func (table *Table) RemovePlayer(player *Player) {
	for i, p := range table.Players {
		if p == player {
			table.Players = append(table.Players[:i], table.Players[i+1:]...)
			return
		}
	}
}
