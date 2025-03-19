package blackjack

import (
	"errors"
	"math/rand/v2"
)

type Rank int

const (
	Ace Rank = iota
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
)

type Suit int

const (
	Spades Suit = iota
	Hearts
	Clubs
	Diamonds
)

// Card is a single card within a deck.
type Card struct {
	Rank Rank
	Suit Suit
}

// Deck is a deck of cards.
type Deck struct {
	Cards []*Card
}

// Shoe contains one or more decks of cards.
type Shoe struct {
	Cards    []*Card
	numDecks int
}

// Hand contains the cards that a player or dealer has been dealt.
type Hand struct {
	Cards       []*Card
	Bet         int
	DoubledDown bool
	Surrendered bool
}

// NewDeck returns a new deck. The cards within the deck are not shuffled.
func NewDeck() *Deck {
	d := &Deck{
		Cards: make([]*Card, 0),
	}
	for _, suite := range []Suit{Spades, Hearts, Clubs, Diamonds} {
		for _, card := range []Rank{Ace, Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King} {
			d.Cards = append(d.Cards, &Card{Rank: card, Suit: suite})
		}
	}
	return d
}

// NewShoe returns a new shoe with numDecks number of decks. The cards within the shoe are not shuffled.
func NewShoe(numDecks int) *Shoe {
	shoe := &Shoe{
		Cards:    make([]*Card, 0),
		numDecks: numDecks,
	}
	shoe.setCards()

	return shoe
}

// NewHand returns a new hand for the dealer or player.
func NewHand() *Hand {
	return &Hand{
		Cards: make([]*Card, 0),
	}
}

// Shuffle shuffles the cards within the shoe.
func (shoe *Shoe) Shuffle() *Shoe {
	rand.Shuffle(len(shoe.Cards), func(i, j int) {
		shoe.Cards[i], shoe.Cards[j] = shoe.Cards[j], shoe.Cards[i]
	})
	return shoe
}

// Deal deals a card from the shoe. If the cutoff for the shoe has been readched,
// then the shoe will be reset with a new set of cards.
func (shoe *Shoe) Deal() *Card {
	if len(shoe.Cards) == 0 {
		shoe.setCards()
	}
	card := shoe.Cards[0]
	shoe.Cards[0] = nil
	shoe.Cards = shoe.Cards[1:]
	return card
}

// setCards adds the cards to the shoe and shuffles them.
func (shoe *Shoe) setCards() {
	for range shoe.numDecks {
		d := NewDeck()
		shoe.Cards = append(shoe.Cards, d.Cards...)
	}
	shoe.Shuffle()

	// 70% to 90% of the deck size for cuttoff
	cuttoff := len(shoe.Cards) * (70 + rand.N(21)) / 100
	shoe.Cards = shoe.Cards[:cuttoff]
}

// DoubleDown doubles down on the hand. The player will receive one more card and the bet will be doubled.
func (hand *Hand) DoubleDown(card *Card) error {
	if hand.DoubledDown {
		return errors.New("hand has already been doubled down")
	}
	hand.Cards = append(hand.Cards, card)
	hand.DoubledDown = true
	hand.Bet *= 2
	return nil
}

// Busted returns true if the hand has busted.
func (hand *Hand) Busted() bool {
	return hand.Score() > 21
}

// HasBlackjack returns true if the hand is a blackjack.
func (hand *Hand) HasBlackjack() bool {
	return len(hand.Cards) == 2 && hand.Score() == 21
}

// HasAce returns true if the hand has an ace.
func (hand *Hand) HasAce() bool {
	for _, card := range hand.Cards {
		if card.Rank == Ace {
			return true
		}
	}
	return false
}

// Score returns the score of the hand. Aces are worth 11 points unless the hand would bust.
func (hand *Hand) Score() int {
	score := 0
	numAces := 0
	for _, card := range hand.Cards {
		switch card.Rank {
		case Ace:
			numAces++
			score += 11
		case Jack, Queen, King:
			score += 10
		default:
			score += int(card.Rank) + 1
		}
	}
	for numAces > 0 && score > 21 {
		score -= 10
		numAces--
	}
	return score
}

// GetValue returns the value of the card. Aces are worth 11 points.
func (card *Card) GetValue() int {
	switch card.Rank {
	case Ace:
		return 11
	case Jack, Queen, King:
		return 10
	default:
		return int(card.Rank) + 1
	}
}

// String returns a string representation of the card.
func (c *Card) String() string {
	return c.Rank.String() + " of " + c.Suit.String()
}

// String returns a string representation of the rank.
func (r Rank) String() string {
	switch r {
	case Ace:
		return "Ace"
	case Two:
		return "Two"
	case Three:
		return "Three"
	case Four:
		return "Four"
	case Five:
		return "Five"
	case Six:
		return "Six"
	case Seven:
		return "Seven"
	case Eight:
		return "Eight"
	case Nine:
		return "Nine"
	case Ten:
		return "Ten"
	case Jack:
		return "Jack"
	case Queen:
		return "Queen"
	case King:
		return "King"
	default:
		return "Unknown"
	}
}

// String returns a string representation of the suit.
func (s Suit) String() string {
	switch s {
	case Spades:
		return "Spades"
	case Hearts:
		return "Hearts"
	case Clubs:
		return "Clubs"
	case Diamonds:
		return "Diamonds"
	default:
		return "Unknown"
	}
}
