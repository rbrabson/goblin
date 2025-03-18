package blackjack

import (
	"errors"
	"fmt"
)

type Player struct {
	Hands        []*Hand
	Table        *Table
	HasInsurance bool
}

// NewPlayer returns a new player with a single hand.
func NewPlayer(initialBet int) *Player {
	return &Player{
		Hands: []*Hand{
			NewHand(),
		},
	}
}

// Split splits a player's hand into two hands.
func (player *Player) Split(hand *Hand) error {
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
