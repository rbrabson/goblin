package blackjack

// Dealer is the dealer at a blackjack table.
type Dealer struct {
	Hand
	Table *Table
}

// PlayHand deals cards to the dealer until the dealer's score is 17 or higher.
func (dealer *Dealer) PlayHand() {
	for dealer.Hand.Score() < 17 {
		card := dealer.Table.Shoe.Deal()
		dealer.Hand.Cards = append(dealer.Hand.Cards, card)
		if dealer.HasSoft17() {
			if dealer.Table.Config.HoldOn17 == H17 {
				break
			}
		}
	}
}

// HasSoft17 returns true if the dealer has a soft 17. A soft 17 is when
// the the dealer has a score of 17 and the hand contains at least one ace.
func (dealer *Dealer) HasSoft17() bool {
	return dealer.Hand.Score() == 17 && dealer.Hand.HasAce()
}
