package slots

import (
	"fmt"
	"log/slog"
)

const (
	DummyGuildID = "000000000000000000"
)

// SlotMachine represents a slot machine with a lookup table, payout table, and symbol table.
type SlotMachine struct {
	LookupTable LookupTable
	PayoutTable PayoutTable
	Symbols     *SymbolTable
}

// GetSlotMachine returns a new instance of the SlotMachine.
func GetSlotMachine() *SlotMachine {
	return newSlotMachine()
}

// newSlotMachine creates a new instance of the SlotMachine with initialized lookup table, payout table, and symbol table.
func newSlotMachine() *SlotMachine {
	slotMachine := &SlotMachine{
		LookupTable: GetLookupTable(),
		PayoutTable: GetPayoutTable(),
		Symbols:     GetSymbolTable(),
	}

	return slotMachine
}

// Spin represents the result of a spin in the slot machine game, including the winning index and the symbols displayed.
// The spin contains multiple rows of symbols, with the winning row indicated by Payline. THe multiple rows are used
// to create the multiple display lines.
type SpinResult struct {
	TopLine    Spin
	Payline    Spin
	BottomLine Spin
	Bet        int
	Payout     int
	Message    string
}

// String returns a string representation of the Spin.
func (s *SpinResult) String() string {
	return fmt.Sprintf("Spin{Payline: %v, TopLine: %v, BottomLine: %v, Bet: %d, Payout: %d}", s.Payline, s.TopLine, s.BottomLine, s.Bet, s.Payout)
}

// Spin simulates a spin of the slot machine with the given bet amount and returns the result of the spin,
// including the payline, previous line, next line, bet amount, and payout amount.
func (sm *SlotMachine) Spin(bet int) *SpinResult {
	paylineIndices, payline := sm.LookupTable.GetPaylineSpin()
	previousIndices, previousLine := sm.LookupTable.GetPreviousSpin(paylineIndices)
	_, nextLine := sm.LookupTable.GetNextSpin(paylineIndices, previousIndices)
	payout, payoutmessage := sm.PayoutTable.GetPayoutAmount(bet, payline)

	spinResult := &SpinResult{
		Payline:    payline,
		BottomLine: previousLine,
		TopLine:    nextLine,
		Bet:        bet,
		Payout:     payout,
		Message:    payoutmessage,
	}

	slog.Debug("slot machine spin result",
		slog.Int("bet", spinResult.Bet),
		slog.Int("payout", spinResult.Payout),
	)

	return spinResult
}
