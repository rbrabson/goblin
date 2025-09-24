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
	Color string `json:"color" bson:"color"`
}

type Symbols struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name    string             `json:"name" bson:"name"`
	GuildID string             `json:"guild_id" bson:"guild_id"`
	Symbols []Symbol           `json:"symbols" bson:"symbols"`
}

func GetSymbols(guildID string) *Symbols {
	// TODO: try to read from the DB
	return newSymbols(guildID)
}

func newSymbols(guildID string) *Symbols {
	symbols := readSymbolsFromFile(guildID)
	// TODO: write to DB
	return symbols
}

func readSymbolsFromFile(guildID string) *Symbols {
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

	symbols := &Symbols{
		GuildID: guildID,
		Name:    symbolsTheme,
	}
	err = json.Unmarshal(bytes, &symbols.Symbols)
	if err != nil {
		slog.Error("failed to unmarshal symbols",
			slog.String("guildID", symbols.GuildID),
			slog.String("name", symbols.Name),
			slog.Any("error", err),
		)
		return nil
	}

	slog.Info("loaded symbols",
		slog.String("guildID", symbols.GuildID),
		slog.Int("count", len(symbols.Symbols)),
	)

	return symbols
}
