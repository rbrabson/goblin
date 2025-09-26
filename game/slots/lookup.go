package slots

import (
	"encoding/json"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

const (
	LOOKUP_TABLE_NAME = "lookup"
	NUM_SPINS         = 3
)

// Slot represents a slot symbol in the lookup table.
type Slot string

// String returns a string representation of the Slot.
func (s Slot) String() string {
	return string(s)
}

// Reel represents a single reel of symbols in the slot machine.
type Reel []Symbol

// String returns a string representation of the Reel.
func (r Reel) String() string {
	sb := strings.Builder{}
	sb.WriteString("Symbols{")
	for i, symbol := range r {
		sb.WriteString(symbol.String())
		if i < len(r)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("}")
	return sb.String()
}

// LookupTable represents the lookup table for a guild, containing the reels of slot symbols.
// The lookup table is used to determine the outcome of spins in the slot machine game.
type LookupTable []Reel

// String returns a string representation of the LookupTable.
func (lt LookupTable) String() string {
	sb := strings.Builder{}
	sb.WriteString("LookupTable{")
	sb.WriteString(", Reels: [")
	for i, reel := range lt {
		sb.WriteString(reel.String())
		if i < len(lt)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")
	sb.WriteString("}")
	return sb.String()
}

// Spin represents a single row of symbols displayed during a spin in the slot machine game.
type Spin []Symbol

// String returns a string representation of the SingleSpin.
func (s Spin) String() string {
	sb := strings.Builder{}
	sb.WriteString("Spin{")
	for i, symbol := range s {
		sb.WriteString(symbol.String())
		if i < len(s)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("}")
	return sb.String()
}

// GetLookupTable retrieves the lookup table for the specified guild.
func GetLookupTable() LookupTable {
	// TODO: try to read from the DB
	return newLookupTable()
}

// newLookupTable creates a new lookup table for the specified guild by reading from a configuration file.
func newLookupTable() LookupTable {
	lookupTable := readLookupTableFromFile()
	return lookupTable
}

// readLookupTableFromFile reads the lookup table from a JSON configuration file.
// The file is expected to be located at DISCORD_CONFIG_DIR/slots/lookuptable/lookup.json
// and contain an array of reels, where each reel is an object with a "Slots" field
// that is an array of slot symbols.
func readLookupTableFromFile() LookupTable {
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	configFileName := filepath.Join(configDir, "slots", "lookuptable", LOOKUP_TABLE_NAME+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		return nil
	}

	reels := [][]Slot{}
	err = json.Unmarshal(bytes, &reels)
	if err != nil {
		return nil
	}

	lookupTable := LookupTable{}
	for range len(reels) {
		reel := make([]Symbol, 0, 64)
		lookupTable = append(lookupTable, reel)
	}

	symbolTable := GetSymbols()
	if symbolTable == nil {
		return nil
	}
	for i, reel := range reels {
		for _, slot := range reel {
			symbol, ok := symbolTable.Symbols[string(slot)]
			if !ok {
				slog.Error("failed to find symbol for slot in lookup table",
					slog.String("file", configFileName),
					slog.String("slot", string(slot)),
				)
				return nil
			}
			lookupTable[i] = append(lookupTable[i], symbol)
		}
	}

	slog.Debug("create new lookup table")

	return lookupTable
}

// GetPaylineSpin selects a random symbol from each reel to create the current spin.
// It returns the indices of the selected symbols and the symbols themselves.
func (lt LookupTable) GetPaylineSpin() ([]int, Spin) {
	currentIndices := make([]int, 0, len(lt))
	for _, reel := range lt {
		randIndex := rand.Int31n(int32(len(reel)))
		currentIndices = append(currentIndices, int(randIndex))
	}
	currentSpin := Spin{}
	for i, reel := range lt {
		currentSpin = append(currentSpin, reel[currentIndices[i]])
	}
	return currentIndices, currentSpin
}

// GetPreviousSpin determines the previous spin based on the current indices.
// It returns the indices of the previous symbols and the symbols themselves.
// The previous symbol for each reel is the first symbol that is different from the current symbol,
func (lt LookupTable) GetPreviousSpin(currentIndices []int) ([]int, Spin) {
	previousSpin := Spin{}
	previousIndices := make([]int, 0, len(lt))
	for i, reel := range lt {
		previousIndex := lt.GetPreviousIndex(reel, currentIndices[i])
		previousSpin = append(previousSpin, reel[previousIndex])
		previousIndices = append(previousIndices, previousIndex)
	}
	return previousIndices, previousSpin
}

// GetPreviousIndex finds the index of the previous symbol in the reel that is different from the current symbol.
// It wraps around to the end of the reel if necessary.
func (lt LookupTable) GetPreviousIndex(reel []Symbol, currentIndex int) int {
	currentSymbol := reel[currentIndex].Name
	previousIndex := currentIndex
	for {
		previousIndex--
		if previousIndex < 0 {
			previousIndex = len(reel) - 1
		}
		if reel[previousIndex].Name != currentSymbol {
			break
		}
	}
	return previousIndex
}

// GetNextSpin determines the next spin based on the current indices.
// It returns the indices of the next symbols and the symbols themselves.
// The next symbol for each reel is the first symbol that is different from the current symbol.
func (lt LookupTable) GetNextSpin(currentIndices []int, previousIndices []int) ([]int, Spin) {
	nextSpin := Spin{}
	nextIndices := make([]int, 0, len(lt))
	for i, reel := range lt {
		nextIndex := lt.GetNextIndex(reel, currentIndices[i], previousIndices[i])
		nextSpin = append(nextSpin, reel[nextIndex])
		nextIndices = append(nextIndices, nextIndex)
	}
	return nextIndices, nextSpin
}

// GetNextIndex finds the index of the next symbol in the reel that is different from the current symbol.
// It wraps around to the beginning of the reel if necessary.
func (lt LookupTable) GetNextIndex(reel []Symbol, currentIndex int, previousIndex int) int {
	currentSymbol := reel[currentIndex].Name
	previousSymbol := reel[previousIndex].Name
	nextIndex := currentIndex
	for {
		nextIndex++
		if nextIndex > len(reel)-1 {
			nextIndex = 0
		}
		if reel[nextIndex].Name != currentSymbol && reel[nextIndex].Name != previousSymbol {
			break
		}
	}
	return nextIndex
}
