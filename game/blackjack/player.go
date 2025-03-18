package blackjack

import (
	"errors"
	"fmt"
)

type Player struct {
	Hands        []*Hand
	Table        *Table
	HasInsurance bool
	currentHand  int
}

// NewPlayer returns a new player with a single hand.
func NewPlayer(initialBet int) *Player {
	return &Player{
		Hands: []*Hand{
			NewHand(),
		},
	}
}

// Hit adds a card to the player's hand.
func (player *Player) Hit() error {
	hand := player.Hands[player.currentHand]
	if hand.Busted() {
		return errors.New("cannot hit on a busted hand")
	}
	newCard := player.Table.Shoe.Deal()
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
	hand := player.Hands[player.currentHand]
	if player.Table.Config.NumSplits == 0 {
		return fmt.Errorf("cannot split")
	}
	if len(player.Hands) > player.Table.Config.NumSplits {
		return fmt.Errorf("cannot split more than %d times", player.Table.Config.NumSplits)
	}
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

// BuyInsurance buys insurance for the player.
func (player *Player) BuyInsurance() error {
	if player.Table.Dealer.Cards[0].Rank != Ace {
		return errors.New("dealer does not have an Ace")
	}
	player.HasInsurance = true
	return nil
}
