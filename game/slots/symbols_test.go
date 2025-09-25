package slots

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestSymbol_String(t *testing.T) {
	tests := []struct {
		name     string
		symbol   Symbol
		expected string
	}{
		{
			name: "basic symbol",
			symbol: Symbol{
				Name:  "Archer Queen",
				Emoji: "<:Archer_Queen:1346562884720922707>",
			},
			expected: "Symbol{Name: Archer Queen, Emoji: <:Archer_Queen:1346562884720922707>}",
		},
		{
			name: "symbol with empty fields",
			symbol: Symbol{
				Name:  "",
				Emoji: "",
			},
			expected: "Symbol{Name: , Emoji: }",
		},
		{
			name: "symbol with special characters",
			symbol: Symbol{
				Name:  "Test Symbol!@#$%",
				Emoji: "🎰",
			},
			expected: "Symbol{Name: Test Symbol!@#$%, Emoji: 🎰}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.symbol.String()
			if result != tt.expected {
				t.Errorf("Symbol.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReel_String(t *testing.T) {
	tests := []struct {
		name     string
		reel     Reel
		expected string
	}{
		{
			name:     "empty reel",
			reel:     Reel{},
			expected: "Symbols{}",
		},
		{
			name: "single symbol reel",
			reel: Reel{
				{Name: "Archer", Emoji: "<:Archer:123>"},
			},
			expected: "Symbols{Symbol{Name: Archer, Emoji: <:Archer:123>}}",
		},
		{
			name: "multiple symbols reel",
			reel: Reel{
				{Name: "Archer", Emoji: "<:Archer:123>"},
				{Name: "Wizard", Emoji: "<:Wizard:456>"},
				{Name: "Barbarian", Emoji: "<:Barbarian:789>"},
			},
			expected: "Symbols{Symbol{Name: Archer, Emoji: <:Archer:123>}, Symbol{Name: Wizard, Emoji: <:Wizard:456>}, Symbol{Name: Barbarian, Emoji: <:Barbarian:789>}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.reel.String()
			if result != tt.expected {
				t.Errorf("Reel.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSymbolTable_String(t *testing.T) {
	objectID := primitive.NewObjectID()

	tests := []struct {
		name        string
		symbolTable SymbolTable
		expected    string
	}{
		{
			name: "empty symbol table",
			symbolTable: SymbolTable{
				ID:      objectID,
				GuildID: "12345",
				Symbols: make(map[string]Symbol),
			},
			expected: "ID: " + objectID.Hex() + ", GuildID: 12345, Symbols: []",
		},
		{
			name: "symbol table with one symbol",
			symbolTable: SymbolTable{
				ID:      objectID,
				GuildID: "12345",
				Symbols: map[string]Symbol{
					"Archer": {Name: "Archer", Emoji: "<:Archer:123>"},
				},
			},
			expected: "ID: " + objectID.Hex() + ", GuildID: 12345, Symbols: [Symbol{Name: Archer, Emoji: <:Archer:123>}]",
		},
		{
			name: "symbol table with multiple symbols (alphabetically sorted)",
			symbolTable: SymbolTable{
				ID:      objectID,
				GuildID: "67890",
				Symbols: map[string]Symbol{
					"Wizard":    {Name: "Wizard", Emoji: "<:Wizard:456>"},
					"Archer":    {Name: "Archer", Emoji: "<:Archer:123>"},
					"Barbarian": {Name: "Barbarian", Emoji: "<:Barbarian:789>"},
				},
			},
			expected: "ID: " + objectID.Hex() + ", GuildID: 67890, Symbols: [Symbol{Name: Archer, Emoji: <:Archer:123>}, Symbol{Name: Barbarian, Emoji: <:Barbarian:789>}, Symbol{Name: Wizard, Emoji: <:Wizard:456>}]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.symbolTable.String()
			if result != tt.expected {
				t.Errorf("SymbolTable.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReadSymbolsFromFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "slots_symbols_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create the nested directory structure
	symbolsDir := filepath.Join(tempDir, "slots", "symbols")
	err = os.MkdirAll(symbolsDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Save original env var and restore it after tests
	originalConfigDir := os.Getenv("DISCORD_CONFIG_DIR")
	defer func() {
		if originalConfigDir == "" {
			os.Unsetenv("DISCORD_CONFIG_DIR")
		} else {
			os.Setenv("DISCORD_CONFIG_DIR", originalConfigDir)
		}
	}()

	// Set the config dir to our temp directory
	os.Setenv("DISCORD_CONFIG_DIR", tempDir)

	tests := []struct {
		name         string
		setupFile    func() error
		guildID      string
		expectNil    bool
		expectSymbol bool
		symbolName   string
	}{
		{
			name: "valid symbols file",
			setupFile: func() error {
				symbols := []Symbol{
					{Name: "Archer Queen", Emoji: "<:Archer_Queen:1346562884720922707>"},
					{Name: "Barbarian King", Emoji: "<:Barbarian_King:1346562986810146826>"},
				}
				data, err := json.Marshal(symbols)
				if err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(symbolsDir, "symbols.json"), data, 0644)
			},
			guildID:      "12345",
			expectNil:    false,
			expectSymbol: true,
			symbolName:   "Archer Queen",
		},
		{
			name: "empty symbols file",
			setupFile: func() error {
				symbols := []Symbol{}
				data, err := json.Marshal(symbols)
				if err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(symbolsDir, "symbols.json"), data, 0644)
			},
			guildID:      "12345",
			expectNil:    false,
			expectSymbol: false,
		},
		{
			name: "invalid JSON file",
			setupFile: func() error {
				return os.WriteFile(filepath.Join(symbolsDir, "symbols.json"), []byte("invalid json"), 0644)
			},
			guildID:   "12345",
			expectNil: true,
		},
		{
			name: "missing file",
			setupFile: func() error {
				// Remove the file if it exists
				os.Remove(filepath.Join(symbolsDir, "symbols.json"))
				return nil
			},
			guildID:   "12345",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the file for this test
			err := tt.setupFile()
			if err != nil {
				t.Fatalf("Failed to setup test file: %v", err)
			}

			result := readSymbolsFromFile(tt.guildID)

			if tt.expectNil {
				if result != nil {
					t.Errorf("readSymbolsFromFile() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Errorf("readSymbolsFromFile() = nil, want non-nil")
				return
			}

			if result.GuildID != tt.guildID {
				t.Errorf("readSymbolsFromFile().GuildID = %v, want %v", result.GuildID, tt.guildID)
			}

			if tt.expectSymbol {
				if _, exists := result.Symbols[tt.symbolName]; !exists {
					t.Errorf("readSymbolsFromFile() missing expected symbol: %v", tt.symbolName)
				}
			}
		})
	}
}

func TestGetSymbols(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "slots_symbols_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create the nested directory structure
	symbolsDir := filepath.Join(tempDir, "slots", "symbols")
	err = os.MkdirAll(symbolsDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Save original env var and restore it after tests
	originalConfigDir := os.Getenv("DISCORD_CONFIG_DIR")
	defer func() {
		if originalConfigDir == "" {
			os.Unsetenv("DISCORD_CONFIG_DIR")
		} else {
			os.Setenv("DISCORD_CONFIG_DIR", originalConfigDir)
		}
	}()

	// Set the config dir to our temp directory
	os.Setenv("DISCORD_CONFIG_DIR", tempDir)

	// Create a valid symbols file
	symbols := []Symbol{
		{Name: "Test Symbol 1", Emoji: "🎰"},
		{Name: "Test Symbol 2", Emoji: "🎲"},
	}
	data, err := json.Marshal(symbols)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(symbolsDir, "symbols.json"), data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		guildID string
	}{
		{
			name:    "get symbols for guild",
			guildID: "12345",
		},
		{
			name:    "get symbols for different guild",
			guildID: "67890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSymbols(tt.guildID)

			if result == nil {
				t.Errorf("GetSymbols() = nil, want non-nil")
				return
			}

			if result.GuildID != tt.guildID {
				t.Errorf("GetSymbols().GuildID = %v, want %v", result.GuildID, tt.guildID)
			}

			// Verify that symbols were loaded
			if len(result.Symbols) == 0 {
				t.Errorf("GetSymbols() returned empty symbols map")
			}

			// Check specific symbols
			if _, exists := result.Symbols["Test Symbol 1"]; !exists {
				t.Errorf("GetSymbols() missing expected symbol: Test Symbol 1")
			}
			if _, exists := result.Symbols["Test Symbol 2"]; !exists {
				t.Errorf("GetSymbols() missing expected symbol: Test Symbol 2")
			}
		})
	}
}

func TestNewSymbols(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "slots_symbols_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create the nested directory structure
	symbolsDir := filepath.Join(tempDir, "slots", "symbols")
	err = os.MkdirAll(symbolsDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Save original env var and restore it after tests
	originalConfigDir := os.Getenv("DISCORD_CONFIG_DIR")
	defer func() {
		if originalConfigDir == "" {
			os.Unsetenv("DISCORD_CONFIG_DIR")
		} else {
			os.Setenv("DISCORD_CONFIG_DIR", originalConfigDir)
		}
	}()

	// Set the config dir to our temp directory
	os.Setenv("DISCORD_CONFIG_DIR", tempDir)

	// Create a valid symbols file
	symbols := []Symbol{
		{Name: "New Symbol 1", Emoji: "⭐"},
		{Name: "New Symbol 2", Emoji: "💎"},
	}
	data, err := json.Marshal(symbols)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(symbolsDir, "symbols.json"), data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	guildID := "test-guild-123"
	result := newSymbols(guildID)

	if result == nil {
		t.Errorf("newSymbols() = nil, want non-nil")
		return
	}

	if result.GuildID != guildID {
		t.Errorf("newSymbols().GuildID = %v, want %v", result.GuildID, guildID)
	}

	// Verify that symbols were loaded
	if len(result.Symbols) != 2 {
		t.Errorf("newSymbols() returned %d symbols, want 2", len(result.Symbols))
	}

	// Check specific symbols
	if symbol, exists := result.Symbols["New Symbol 1"]; !exists {
		t.Errorf("newSymbols() missing expected symbol: New Symbol 1")
	} else if symbol.Emoji != "⭐" {
		t.Errorf("newSymbols() symbol emoji = %v, want ⭐", symbol.Emoji)
	}

	if symbol, exists := result.Symbols["New Symbol 2"]; !exists {
		t.Errorf("newSymbols() missing expected symbol: New Symbol 2")
	} else if symbol.Emoji != "💎" {
		t.Errorf("newSymbols() symbol emoji = %v, want 💎", symbol.Emoji)
	}
}
