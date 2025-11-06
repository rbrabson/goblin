package slots

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rbrabson/goblin/discord"
)

var (
	symbolTable SymbolTable
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

// SymbolTable defines a table of symbols for a specific guild.
type SymbolTable map[string]Symbol

// String returns a string representation of the SymbolTable.
func (st SymbolTable) String() string {
	sb := strings.Builder{}
	symbolNames := make([]string, 0, len(st))
	for name := range st {
		symbolNames = append(symbolNames, name)
	}
	slices.Sort(symbolNames)
	sb.WriteString(", Symbols: [")
	for i, name := range symbolNames {
		symbol := st[name]
		sb.WriteString(symbol.String())
		if i < len(symbolNames)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

// GetSymbolTable retrieves the symbol table for a specific guild.
func GetSymbolTable() SymbolTable {
	if symbolTable == nil {
		symbolTable = newSymbolTable()
	}
	return symbolTable
}

// GetSymbolNames returns a slice of symbol names in the symbol table.
func newSymbolTable() SymbolTable {
	symbols := readSymbolTableFromFile()
	return symbols
}

// readSymbolTableFromFile reads the symbol table from a JSON file.
func readSymbolTableFromFile() SymbolTable {
	configFileName := filepath.Join(discord.DISCORD_CONFIG_DIR, "slots", "symbols", slotsTheme+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		slog.Error("failed to read symbols file",
			slog.String("file", configFileName),
			slog.Any("error", err),
		)
		return nil
	}

	symbols := make([]Symbol, 0)
	err = json.Unmarshal(bytes, &symbols)
	if err != nil {
		slog.Error("failed to unmarshal symbols",
			slog.Any("error", err),
		)
		return nil
	}

	symbolTable := make(SymbolTable)
	for _, symbol := range symbols {
		symbolTable[symbol.Name] = symbol

	}

	slog.Debug("loaded symbols")

	return symbolTable
}
