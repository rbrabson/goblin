package slots

import (
	"fmt"
)

const (
	DummyGuildID = "000000000000000000"
)

type SlotMachine struct {
	LookupTable *LookupTable
	PayoutTable *PayoutTable
	Symbols     *SymbolTable
}

func NewSlotMachine(guildID string) *SlotMachine {
	slotMachine := &SlotMachine{
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
	NextLine     Spin
	Payline      Spin
	PreviousLine Spin
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
	paylineIndices, payline := sm.LookupTable.GetPaylineSpin()
	previousIndices, previousLine := sm.LookupTable.GetPreviousSpin(paylineIndices)
	_, nextLine := sm.LookupTable.GetNextSpin(paylineIndices, previousIndices)
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
