package blackjack

import (
	"errors"
	"fmt"

	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/guild"
)

type Player struct {
	Hands        []*Hand
	table        *Table
	HasInsurance bool
	currentHand  int
	guildMember  *guild.Member
	bankAccount  *bank.Account
}

// NewPlayer returns a new player with a single hand.
func NewPlayer(guildID string, memberID string) *Player {
	return &Player{
		Hands: []*Hand{
			NewHand(),
		},
		guildMember: guild.GetMember(guildID, memberID),
		bankAccount: bank.GetAccount(guildID, memberID),
	}
}

// Hit adds a card to the player's hand.
func (player *Player) Hit() error {
	if player.currentHand+1 >= len(player.Hands) {
		return errors.New("cannot hit after playing last hand")
	}
	hand := player.Hands[player.currentHand]
	if hand.Busted() {
		return errors.New("cannot hit on a busted hand")
	}
	newCard := player.table.Shoe.Deal()
	hand.Cards = append(hand.Cards, newCard)
	if hand.Busted() {
		return errors.New("hand busted")
	}
	return nil
}

// Stand has the player take no more actions on the current hand. If the player has
// more hands to play, then the next hand will be played. If the player has no more
// hands to play, then the player is done.
func (player *Player) Stand() bool {
	player.currentHand++
	return player.currentHand+1 >= len(player.Hands)
}

// Split splits a player's hand into two hands.
func (player *Player) Split() error {
	if player.table.Config.NumSplits == 0 {
		return fmt.Errorf("cannot split")
	}
	if len(player.Hands) > player.table.Config.NumSplits {
		return fmt.Errorf("cannot split more than %d times", player.table.Config.NumSplits)
	}
	if player.currentHand+1 >= len(player.Hands) {
		return errors.New("cannot split after playing last hand")
	}
	hand := player.Hands[player.currentHand]
	if hand.Cards[0].Rank != hand.Cards[1].Rank {
		return fmt.Errorf("cannot split %s and %s", hand.Cards[0].Rank, hand.Cards[1].Rank)
	}

	newHand := NewHand()
	newHand.Bet = hand.Bet
	newHand.Cards = append(newHand.Cards, hand.Cards[1])
	hand.Cards = hand.Cards[:1]
	player.Hands = append(player.Hands, newHand)
	return nil
}

// Surrender surrenders the player's hand. The player loses the hand, but only a portion
// of the bet is lost.
func (player *Player) Surrender() error {
	if len(player.Hands) > 1 {
		return errors.New("cannot surrender with multiple hands")
	}
	if player.currentHand+1 >= len(player.Hands) {
		return errors.New("can't surrender after playing last hand")
	}
	hand := player.Hands[0]
	if hand.Busted() {
		return errors.New("cannot surrender a busted hand")
	}

	player.Hands[0].Surrendered = true
	player.currentHand++
	return nil
}

// BuyInsurance buys insurance for the player.
func (player *Player) BuyInsurance() error {
	if player.table.Dealer.Cards[0].Rank != Ace {
		return errors.New("dealer does not have an Ace")
	}
	player.HasInsurance = true
	return nil
}
