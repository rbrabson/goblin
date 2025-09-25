package slots

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	SYMBOLS_FILE_NAME = "symbols"
	// SYMBOLS_FILE_NAME = "pt"
)

// Symbol represents a slot symbol with a name and an emoji.
type Symbol struct {
	Name  string `json:"name" bson:"name"`
	Emoji string `json:"emoji" bson:"emoji"`
}

// String returns a string representation of the Symbol.
func (s *Symbol) String() string {
	sb := strings.Builder{}
	sb.WriteString("Symbol{")
	sb.WriteString("Name: " + s.Name)
	sb.WriteString(", Emoji: " + s.Emoji)
	sb.WriteString("}")

	return sb.String()
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

// SymbolTable defines a table of symbols for a specific guild.
type SymbolTable struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	GuildID string             `json:"guild_id" bson:"guild_id"`
	Symbols map[string]Symbol  `json:"symbols" bson:"symbols"`
}

// String returns a string representation of the SymbolTable.
func (st *SymbolTable) String() string {
	sb := strings.Builder{}
	sb.WriteString("ID: " + st.ID.Hex())
	sb.WriteString(", GuildID: " + st.GuildID)
	symbolNames := make([]string, 0, len(st.Symbols))
	for name := range st.Symbols {
		symbolNames = append(symbolNames, name)
	}
	slices.Sort(symbolNames)
	sb.WriteString(", Symbols: [")
	for i, name := range symbolNames {
		symbol := st.Symbols[name]
		sb.WriteString(symbol.String())
		if i < len(symbolNames)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

// GetSymbols retrieves the symbol table for a specific guild.
func GetSymbols(guildID string) *SymbolTable {
	// TODO: try to read from the DB
	return newSymbols(guildID)
}

// GetSymbolNames returns a slice of symbol names in the symbol table.
func newSymbols(guildID string) *SymbolTable {
	symbols := readSymbolsFromFile(guildID)
	// TODO: write to DB
	return symbols
}

// readSymbolsFromFile reads the symbol table from a JSON file.
func readSymbolsFromFile(guildID string) *SymbolTable {
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	configFileName := filepath.Join(configDir, "slots", "symbols", SYMBOLS_FILE_NAME+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		slog.Error("failed to read symbols file",
			slog.String("guildID", guildID),
			slog.String("file", configFileName),
			slog.Any("error", err),
		)
		return nil
	}

	symbolTable := &SymbolTable{
		GuildID: guildID,
		Symbols: make(map[string]Symbol),
	}
	symbols := &[]Symbol{}
	err = json.Unmarshal(bytes, symbols)
	if err != nil {
		slog.Error("failed to unmarshal symbols",
			slog.String("guildID", symbolTable.GuildID),
			slog.Any("error", err),
		)
		return nil
	}

	for _, symbol := range *symbols {
		symbolTable.Symbols[symbol.Name] = symbol

	}

	slog.Debug("loaded symbols",
		slog.String("guildID", symbolTable.GuildID),
	)

	return symbolTable
}
