package slots

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
)

const (
	LOOKUP_TABLE_NAME = "lookup"
	NUM_SPINS         = 10
)

type Slot string

type LookupTable struct {
	GuildID string     `json:"guild_id"`
	Reels   [][]Symbol `json:"reels"`
}

type Spin struct {
	WinIndex int
	Spins    [][]Symbol
}

func GetLookupTable(guildID string) *LookupTable {
	// TODO: try to read from the DB
	return newLookupTable(guildID)
}

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
		Reels:   make([][]Symbol, 0, len(*reels)),
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

func (lt *LookupTable) Spin() *Spin {
	spin := &Spin{
		Spins: make([][]Symbol, 0, NUM_SPINS),
	}

	currentIndices, currentSpin := lt.GetCurrentSpin()
	fmt.Println("Spin:", currentSpin)
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
