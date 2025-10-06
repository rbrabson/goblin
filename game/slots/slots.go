package slots

import (
	rslots "github.com/rbrabson/slots"
)

const (
	DummyGuildID = "000000000000000000"
)

// SlotMachine represents a slot machine with a lookup table, payout table, and symbol table.
type SlotMachine struct {
	slotMachine rslots.SlotMachine
	symbols     SymbolTable
}

// GetSlotMachine returns a new instance of the SlotMachine.
func GetSlotMachine() *SlotMachine {
	return newSlotMachine()
}

// newSlotMachine creates a new instance of the SlotMachine with initialized lookup table, payout table, and symbol table.
func newSlotMachine() *SlotMachine {
	slotMachine := &SlotMachine{
		slotMachine: *rslots.NewSlotMachine(
			rslots.WithLookupTable(GetLookupTable()),
			rslots.WithPayoutTable(GetPayoutTable()),
		),
		symbols: GetSymbolTable(),
	}

	return slotMachine
}

// Spin spins the slot machine with the given bet and returns the result.
func (sm *SlotMachine) Spin(bet int) *rslots.SpinResult {
	return sm.slotMachine.Spin(bet)
}
