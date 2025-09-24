package slots

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	LOOKUP_TABLE_NAME = "lookup"
)

type Slot string

type LookupTable struct {
	GuildID string     `json:"guild_id"`
	Name    string     `json:"name"`
	Reels   [][]Symbol `json:"reels"`
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
		Name:    LOOKUP_TABLE_NAME,
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

	slog.Info("create new lookup table",
		slog.String("guildID", lookupTable.GuildID),
		slog.String("theme", lookupTable.Name),
	)

	return lookupTable
}
