package slots

import (
	"fmt"
)

const (
	DummyGuildID = "000000000000000000"
)

type SlotMachine struct {
	Reels       []Reel
	LookupTable *LookupTable
	PayoutTable *PayoutTable
	Symbols     *SymbolTable
}

func NewSlotMachine(guildID string) *SlotMachine {
	slotMachine := &SlotMachine{
		Reels:       []Reel{},
		LookupTable: GetLookupTable(guildID),
		PayoutTable: GetPayoutTable(guildID),
		Symbols:     GetSymbols(guildID),
	}

	return slotMachine
}

// Spin represents the result of a spin in the slot machine game, including the winning index and the symbols displayed.
// The spin contains multiple rows of symbols, with the winning row indicated by Payline. THe multiple rows are used
// to create the multiple display lines.
type SpinResult struct {
	NextLine     SingleSpin
	Payline      SingleSpin
	PreviousLine SingleSpin
	Bet          int
	Payout       int
}

// String returns a string representation of the Spin.
func (s *SpinResult) String() string {
	return fmt.Sprintf("Spin{Payline: %v, NextLine: %v, PreviousLine: %v, Bet: %d, Payout: %d}", s.Payline, s.NextLine, s.PreviousLine, s.Bet, s.Payout)
}

// Spin simulates a spin of the slot machine with the given bet amount and returns the result of the spin,
// including the payline, previous line, next line, bet amount, and payout amount.
func (sm *SlotMachine) Spin(bet int) *SpinResult {
	paylineIndices, payline := sm.LookupTable.GetCurrentSpin()
	_, previousLine := sm.LookupTable.GetPreviousSpin(paylineIndices)
	_, nextLine := sm.LookupTable.GetNextSpin(paylineIndices)
	payout := sm.PayoutTable.GetPayoutAmount(bet, payline)

	spinResult := &SpinResult{
		Payline:      payline,
		PreviousLine: previousLine,
		NextLine:     nextLine,
		Bet:          bet,
		Payout:       payout,
	}

	return spinResult
}

/*
    Bonus: special features.
	-- Free Spins
	   - Scatter of at least three designated symbols
	Number of wheels
	-- 3 wheels
	-- figure out number of symbols per wheel
	Config
	-- Payout Percentage
	   -- 75% minimum in Nevada; 83% in New Jersey, etc.
	-- Frequency of Payout
	-- Amount of Payout
	-- Amount of Bet
	-- Jackpot Payout
	-- Par is the number of spins needed to reach the payouot percentage
	-- Pay Table
	   -- Payouts for combination of symbols
	-- PayLine
	   -- Line (straight or zizag) across the wheels that determines a win

	-- Paytable
	   -- Paylines
	   -- Reward Probabilities
	   -- Winning Combinations


	-- Example:
	   -- 13 possible payoutts ranging from 1:1 to 2400:1
	      -- 1:1 comes out every 8 plays
		  -- 5:1 comes out every 33 plays
		  -- 2:1 comes out ever 600 plays
		  -- 80:1 comes out every 219 plays
		  -- 150:1 comes every 6,241 plays
		  -- 2400:1 comes every 262,144 plays
	-- RTP (Return to Player)


	Allow pay tables to. be configured, and different pay tables selected by the admin

*/

/*
Let's take a closer look at the individual parts of the online slot pay table:

    Paylines: Paylines mark the patterns on the reels that the symbols need to align with to generate a win.
	          Old-school slots became famous for the single horizontal payline, but modern online slots have
			  multiple paylines and various patterns and directions. Pay tables will usually illustrate the
			  payline patterns across the game grid.

    Bet Limits: Are the game's minimum and maximum bets, including the required bet for the jackpot in the
	            case of a progressive slot. When playing an online slot, you'll often only be able to win certain
				prizes if you bet a specific amount on a spin. For example, the jackpot is usually only possible
				to win if you place a higher bet.

    Paying Symbols: Are all slot game symbols that can appear on the reels during gameplay, including their
	            possible winning combinations. Pay tables are the best place to see every slot symbol in the game
				and what each means to your chances of winning a prize. Symbol images are shown alongside coin
				amounts or multipliers you can win if those combinations land on the reels. As symbols are so varied
				in the modern online slot realm, it is important to check the pay table to know what each symbol means.

    Special Symbols: Examples of these would be the Wild symbol and Scatter symbol, which will trigger unique features
	                 and wins. A pay table will explain how these are activated and how they function within the slot
					 machine's gameplay.

    Features: Pay tables provide the player valuable information about features available on the slot game, such as the
	          bonus rounds, Multipliers, Free Spins, auto-spins, etc. In terms of auto spin, this is an essential feature
			  on modern online slots that lets you pre-set the amount of spins. It speeds up gameplay and makes the game
			  more efficient. However, it is imperative to look at the payable carefully and learn how to use the auto-spin
			  feature safely, while controlling your budget and gambling responsibly.

    Jackpots: Even if you are playing for fun, all online slot gamblers dream of winning the jackpot. What's better than
	          spinning the reels on a single bet and winning a sensational sum of money? A pay table shows information about jackpot
	          payouts and required bets, including progressive jackpot games. Imagine the disappointment of thinking you had won a jackpot
	          and then discovering that you didn't place the qualifying bet amount because you forgot to check the pay table.

   When you read a slot pay table, get familiar with these parts to give yourself more enjoyment when playing the game.
*/

/*
   When spinning, show three lines, with the middle line being the payline. Use a `>` character to point to this. In the lookup table,
   then get the values from above or below to be displayed, as long as the value does not match the selected value.
*/
