package slots

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	LOOKUP_TABLE_NAME = "lookup"
	NUM_SPINS         = 10
)

// Slot represents a slot symbol in the lookup table.
type Slot string

// String returns a string representation of the Slot.
func (s Slot) String() string {
	return string(s)
}

// LookupTable represents the lookup table for a guild, containing the reels of slot symbols.
// The lookup table is used to determine the outcome of spins in the slot machine game.
type LookupTable struct {
	GuildID string `json:"guild_id"`
	Reels   []Reel `json:"reels"`
}

// String returns a string representation of the LookupTable.
func (lt *LookupTable) String() string {
	sb := strings.Builder{}
	sb.WriteString("LookupTable{")
	sb.WriteString("GuildID: " + lt.GuildID)
	sb.WriteString(", Reels: [")
	for i, reel := range lt.Reels {
		sb.WriteString(reel.String())
		if i < len(lt.Reels)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")
	sb.WriteString("}")
	return sb.String()
}

// Spins represents multiple rows of symbols displayed during a spin in the slot machine game.
type Spins []Reel

// String returns a string representation of the Spins.
func (s Spins) String() string {
	sb := strings.Builder{}
	sb.WriteString("Spins{")
	for i, spin := range s {
		sb.WriteString(spin.String())
		if i < len(s)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("}")
	return sb.String()
}

// Spin represents the result of a spin in the slot machine game, including the winning index and the symbols displayed.
// The spin contains multiple rows of symbols, with the winning row indicated by WinIndex. THe multiple rows are used
// to create an animation effect when displaying the spin result.
type Spin struct {
	WinIndex int
	Spins    Spins
}

// String returns a string representation of the Spin.
func (s *Spin) String() string {
	return fmt.Sprintf("Spin{WinIndex: %d, Spins: %v}", s.WinIndex, s.Spins)
}

// GetLookupTable retrieves the lookup table for the specified guild.
func GetLookupTable(guildID string) *LookupTable {
	// TODO: try to read from the DB
	return newLookupTable(guildID)
}

// newLookupTable creates a new lookup table for the specified guild by reading from a configuration file.
func newLookupTable(guildID string) *LookupTable {
	lookupTable := readLookupTableFromFile(guildID)
	if lookupTable == nil {
		slog.Error("failed to create lookup table",
			slog.String("guildID", guildID),
		)
		return nil
	}
	// TODO: write lookup table to the DB
	return lookupTable
}

// readLookupTableFromFile reads the lookup table from a JSON configuration file.
// The file is expected to be located at DISCORD_CONFIG_DIR/slots/lookuptable/lookup.json
// and contain an array of reels, where each reel is an object with a "Slots" field
// that is an array of slot symbols.
func readLookupTableFromFile(guildID string) *LookupTable {
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	configFileName := filepath.Join(configDir, "slots", "lookuptable", LOOKUP_TABLE_NAME+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		return nil
	}

	reels := &[][]Slot{}
	err = json.Unmarshal(bytes, reels)
	if err != nil {
		return nil
	}

	lookupTable := &LookupTable{
		GuildID: guildID,
		Reels:   make([]Reel, 0, len(*reels)),
	}
	for range len(*reels) {
		reels := make([]Symbol, 0, 64)
		lookupTable.Reels = append(lookupTable.Reels, reels)
	}

	symbolTable := GetSymbols(guildID)
	if symbolTable == nil {
		return nil
	}
	for i, reel := range *reels {
		for _, slot := range reel {
			symbol, ok := symbolTable.Symbols[string(slot)]
			if !ok {
				slog.Error("failed to find symbol for slot in lookup table",
					slog.String("guildID", guildID),
					slog.String("file", configFileName),
					slog.String("slot", string(slot)),
				)
				return nil
			}
			lookupTable.Reels[i] = append(lookupTable.Reels[i], symbol)
		}
	}

	slog.Debug("create new lookup table",
		slog.String("guildID", lookupTable.GuildID),
	)

	return lookupTable
}

// Spin generates a new spin result using the lookup table.
// It selects a random current spin, then determines the previous and next spins
// to create an animation effect. The winning index is set to the second-to-last spin.
func (lt *LookupTable) Spin() *Spin {
	spin := &Spin{
		Spins: make([]Reel, 0, NUM_SPINS),
	}

	currentIndices, currentSpin := lt.GetCurrentSpin()
	_, previousSpin := lt.GetPreviousSpin(currentIndices)
	spin.Spins = append(spin.Spins, previousSpin)
	spin.Spins = append(spin.Spins, currentSpin)

	nextIndices := currentIndices
	var nextSpin []Symbol
	for range NUM_SPINS - 2 {
		nextIndices, nextSpin = lt.GetNextSpin(nextIndices)
		spin.Spins = append(spin.Spins, nextSpin)
	}
	slices.Reverse(spin.Spins)
	spin.WinIndex = NUM_SPINS - 2

	return spin
}

// GetCurrentSpin selects a random symbol from each reel to create the current spin.
// It returns the indices of the selected symbols and the symbols themselves.
func (lt *LookupTable) GetCurrentSpin() ([]int, []Symbol) {
	currentIndices := make([]int, 0, len(lt.Reels))
	for _, reel := range lt.Reels {
		randIndex := rand.Int31n(int32(len(reel)))
		currentIndices = append(currentIndices, int(randIndex))
	}
	currentSpin := make([]Symbol, 0, len(lt.Reels))
	for i, reel := range lt.Reels {
		currentSpin = append(currentSpin, reel[currentIndices[i]])
	}
	return currentIndices, currentSpin
}

// GetPreviousSpin determines the previous spin based on the current indices.
// It returns the indices of the previous symbols and the symbols themselves.
// The previous symbol for each reel is the first symbol that is different from the current symbol,
func (lt *LookupTable) GetPreviousSpin(currentIndices []int) ([]int, []Symbol) {
	previousSpin := make([]Symbol, 0, len(lt.Reels))
	previousIndices := make([]int, 0, len(lt.Reels))
	for i, reel := range lt.Reels {
		previousIndex := lt.GetPreviousIndex(reel, currentIndices[i])
		previousSpin = append(previousSpin, reel[previousIndex])
		previousIndices = append(previousIndices, previousIndex)
	}
	return previousIndices, previousSpin
}

// GetPreviousIndex finds the index of the previous symbol in the reel that is different from the current symbol.
// It wraps around to the end of the reel if necessary.
func (lt *LookupTable) GetPreviousIndex(reel []Symbol, currentIndex int) int {
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
func (lt *LookupTable) GetNextSpin(currentIndices []int) ([]int, []Symbol) {
	nextSpin := make([]Symbol, 0, len(lt.Reels))
	nextIndices := make([]int, 0, len(lt.Reels))
	for i, reel := range lt.Reels {
		nextIndex := lt.GetNextIndex(reel, currentIndices[i])
		nextSpin = append(nextSpin, reel[nextIndex])
		nextIndices = append(nextIndices, nextIndex)
	}
	return nextIndices, nextSpin
}

// GetNextIndex finds the index of the next symbol in the reel that is different from the current symbol.
// It wraps around to the beginning of the reel if necessary.
func (lt *LookupTable) GetNextIndex(reel []Symbol, currentIndex int) int {
	currentSymbol := reel[currentIndex].Name
	nextIndex := currentIndex
	for {
		nextIndex++
		if nextIndex > len(reel)-1 {
			nextIndex = 0
		}
		if reel[nextIndex].Name != currentSymbol {
			break
		}
	}
	return nextIndex
}
