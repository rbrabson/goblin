package slots

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	rslots "github.com/rbrabson/slots"
)

const (
	LOOKUP_TABLE_NAME = "lookup"
	NUM_SPINS         = 3
)

// GetLookupTable retrieves the lookup table for the specified guild.
func GetLookupTable() rslots.LookupTable {
	lookupTable := newLookupTable()
	return lookupTable
}

// newLookupTable creates a new lookup table for the specified guild by reading from a configuration file.
func newLookupTable() rslots.LookupTable {
	lookupTable := readLookupTableFromFile()
	return lookupTable
}

// readLookupTableFromFile reads the lookup table from a JSON configuration file.
// The file is expected to be located at DISCORD_CONFIG_DIR/slots/lookuptable/lookup.json
// and contain an array of reels, where each reel is an object with a "Slots" field
// that is an array of slot symbols.
func readLookupTableFromFile() rslots.LookupTable {
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	configFileName := filepath.Join(configDir, "slots", "lookuptable", LOOKUP_TABLE_NAME+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		return nil
	}

	var lookupTable rslots.LookupTable
	err = json.Unmarshal(bytes, &lookupTable)
	if err != nil {
		return nil
	}

	slog.Debug("create new lookup table")

	return lookupTable
}
