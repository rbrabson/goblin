package slots

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Symbol struct {
	Name  string `json:"name" bson:"name"`
	Emoji string `json:"emoji" bson:"emoji"`
}

type SymbolTable struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	GuildID string             `json:"guild_id" bson:"guild_id"`
	Name    string             `json:"name" bson:"name"`
	Symbols map[string]Symbol  `json:"symbols" bson:"symbols"`
}

func GetSymbols(guildID string) *SymbolTable {
	// TODO: try to read from the DB
	return newSymbols(guildID)
}

func newSymbols(guildID string) *SymbolTable {
	symbols := readSymbolsFromFile(guildID)
	// TODO: write to DB
	return symbols
}

func readSymbolsFromFile(guildID string) *SymbolTable {
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	symbolsTheme := os.Getenv("DISCORD_DEFAULT_THEME")
	configFileName := filepath.Join(configDir, "slots", "symbols", symbolsTheme+".json")
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
		Name:    symbolsTheme,
		Symbols: make(map[string]Symbol),
	}
	symbols := &[]Symbol{}
	err = json.Unmarshal(bytes, symbols)
	if err != nil {
		slog.Error("failed to unmarshal symbols",
			slog.String("guildID", symbolTable.GuildID),
			slog.String("name", symbolTable.Name),
			slog.Any("error", err),
		)
		return nil
	}

	for _, symbol := range *symbols {
		symbolTable.Symbols[symbol.Name] = symbol

	}

	slog.Info("loaded symbols",
		slog.String("guildID", symbolTable.GuildID),
		slog.String("name", symbolTable.Name),
	)

	return symbolTable
}
