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

type Reel []Slot

type LookupTable struct {
	GuildID string `json:"guild_id"`
	Name    string `json:"name"`
	Reels   []Reel `json:"reels"`
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
	// write lookup table to the DB
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
		slog.Error("failed to read default theme",
			slog.String("guildID", guildID),
			slog.String("file", configFileName),
			slog.Any("error", err),
		)
		return nil
	}

	lookupTable := &LookupTable{
		GuildID: guildID,
		Name:    LOOKUP_TABLE_NAME,
	}
	// reels := &[][]Slot{}
	err = json.Unmarshal(bytes, &lookupTable.Reels)
	if err != nil {
		slog.Error("failed to unmarshal lookup table",
			slog.String("guildID", guildID),
			slog.String("file", configFileName),
			slog.String("data", string(bytes)),
			slog.Any("error", err),
		)
		return nil
	}

	slog.Info("create new lookup table",
		slog.String("guildID", lookupTable.GuildID),
		slog.String("theme", lookupTable.Name),
	)

	return lookupTable
}
