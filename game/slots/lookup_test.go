package slots

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSlot_String(t *testing.T) {
	tests := []struct {
		name     string
		slot     Slot
		expected string
	}{
		{
			name:     "simple slot",
			slot:     Slot("Archer"),
			expected: "Archer",
		},
		{
			name:     "slot with spaces",
			slot:     Slot("Archer Queen"),
			expected: "Archer Queen",
		},
		{
			name:     "empty slot",
			slot:     Slot(""),
			expected: "",
		},
		{
			name:     "slot with special characters",
			slot:     Slot("Test!@#$%"),
			expected: "Test!@#$%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.slot.String()
			if result != tt.expected {
				t.Errorf("Slot.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLookupTable_String(t *testing.T) {
	tests := []struct {
		name        string
		lookupTable LookupTable
		expected    string
	}{
		{
			name: "empty lookup table",
			lookupTable: LookupTable{
				GuildID: "12345",
				Reels:   []Reel{},
			},
			expected: "LookupTable{GuildID: 12345, Reels: []}",
		},
		{
			name: "lookup table with one reel",
			lookupTable: LookupTable{
				GuildID: "12345",
				Reels: []Reel{
					{{Name: "Archer", Emoji: "<:Archer:123>"}},
				},
			},
			expected: "LookupTable{GuildID: 12345, Reels: [Symbols{Symbol{Name: Archer, Emoji: <:Archer:123>}}]}",
		},
		{
			name: "lookup table with multiple reels",
			lookupTable: LookupTable{
				GuildID: "67890",
				Reels: []Reel{
					{{Name: "Archer", Emoji: "<:Archer:123>"}},
					{{Name: "Wizard", Emoji: "<:Wizard:456>"}, {Name: "Spell", Emoji: "<:Spell:789>"}},
				},
			},
			expected: "LookupTable{GuildID: 67890, Reels: [Symbols{Symbol{Name: Archer, Emoji: <:Archer:123>}}, Symbols{Symbol{Name: Wizard, Emoji: <:Wizard:456>}, Symbol{Name: Spell, Emoji: <:Spell:789>}}]}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.lookupTable.String()
			if result != tt.expected {
				t.Errorf("LookupTable.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSingleSpin_String(t *testing.T) {
	tests := []struct {
		name       string
		singleSpin SingleSpin
		expected   string
	}{
		{
			name:       "empty single spin",
			singleSpin: SingleSpin{},
			expected:   "Spin{}",
		},
		{
			name: "single spin with one symbol",
			singleSpin: SingleSpin{
				{Name: "Archer", Emoji: "<:Archer:123>"},
			},
			expected: "Spin{Symbol{Name: Archer, Emoji: <:Archer:123>}}",
		},
		{
			name: "single spin with multiple symbols",
			singleSpin: SingleSpin{
				{Name: "Archer", Emoji: "<:Archer:123>"},
				{Name: "Wizard", Emoji: "<:Wizard:456>"},
				{Name: "Spell", Emoji: "<:Spell:789>"},
			},
			expected: "Spin{Symbol{Name: Archer, Emoji: <:Archer:123>}, Symbol{Name: Wizard, Emoji: <:Wizard:456>}, Symbol{Name: Spell, Emoji: <:Spell:789>}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.singleSpin.String()
			if result != tt.expected {
				t.Errorf("SingleSpin.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSpins_String(t *testing.T) {
	tests := []struct {
		name     string
		spins    Spins
		expected string
	}{
		{
			name:     "empty spins",
			spins:    Spins{},
			expected: "Spins{}",
		},
		{
			name: "spins with one spin",
			spins: Spins{
				{{Name: "Archer", Emoji: "<:Archer:123>"}},
			},
			expected: "Spins{{Symbol{Name: Archer, Emoji: <:Archer:123>}}}",
		},
		{
			name: "spins with multiple spins",
			spins: Spins{
				{{Name: "Archer", Emoji: "<:Archer:123>"}},
				{{Name: "Wizard", Emoji: "<:Wizard:456>"}, {Name: "Spell", Emoji: "<:Spell:789>"}},
			},
			expected: "Spins{{Symbol{Name: Archer, Emoji: <:Archer:123>}}, {Symbol{Name: Wizard, Emoji: <:Wizard:456>}, Symbol{Name: Spell, Emoji: <:Spell:789>}}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spins.String()
			if result != tt.expected {
				t.Errorf("Spins.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSpin_String(t *testing.T) {
	tests := []struct {
		name     string
		spin     Spin
		expected string
	}{
		{
			name: "basic spin",
			spin: Spin{
				WinIndex: 5,
				Spins: Spins{
					{{Name: "Archer", Emoji: "<:Archer:123>"}},
					{{Name: "Wizard", Emoji: "<:Wizard:456>"}},
				},
			},
			expected: "Spin{WinIndex: 5, Spins: Spins{{Symbol{Name: Archer, Emoji: <:Archer:123>}}, {Symbol{Name: Wizard, Emoji: <:Wizard:456>}}}}",
		},
		{
			name: "spin with zero win index",
			spin: Spin{
				WinIndex: 0,
				Spins:    Spins{},
			},
			expected: "Spin{WinIndex: 0, Spins: Spins{}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spin.String()
			if result != tt.expected {
				t.Errorf("Spin.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReadLookupTableFromFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "slots_lookup_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create the nested directory structure
	lookupDir := filepath.Join(tempDir, "slots", "lookuptable")
	symbolsDir := filepath.Join(tempDir, "slots", "symbols")
	err = os.MkdirAll(lookupDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
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

	// Create symbols file first (required by lookup table)
	symbols := []Symbol{
		{Name: "Archer", Emoji: "<:Archer:123>"},
		{Name: "Wizard", Emoji: "<:Wizard:456>"},
		{Name: "Spell", Emoji: "<:Spell:789>"},
	}
	symbolsData, err := json.Marshal(symbols)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(symbolsDir, "symbols.json"), symbolsData, 0644)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		setupFile  func() error
		guildID    string
		expectNil  bool
		expectReel bool
		reelCount  int
	}{
		{
			name: "valid lookup table file",
			setupFile: func() error {
				lookupData := [][]Slot{
					{"Archer", "Wizard"},
					{"Spell", "Wizard", "Archer"},
				}
				data, err := json.Marshal(lookupData)
				if err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(lookupDir, "lookup.json"), data, 0644)
			},
			guildID:    "12345",
			expectNil:  false,
			expectReel: true,
			reelCount:  2,
		},
		{
			name: "empty lookup table file",
			setupFile: func() error {
				lookupData := [][]Slot{}
				data, err := json.Marshal(lookupData)
				if err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(lookupDir, "lookup.json"), data, 0644)
			},
			guildID:    "12345",
			expectNil:  false,
			expectReel: false,
			reelCount:  0,
		},
		{
			name: "invalid JSON file",
			setupFile: func() error {
				return os.WriteFile(filepath.Join(lookupDir, "lookup.json"), []byte("invalid json"), 0644)
			},
			guildID:   "12345",
			expectNil: true,
		},
		{
			name: "missing file",
			setupFile: func() error {
				// Remove the file if it exists
				os.Remove(filepath.Join(lookupDir, "lookup.json"))
				return nil
			},
			guildID:   "12345",
			expectNil: true,
		},
		{
			name: "unknown symbol in lookup table",
			setupFile: func() error {
				lookupData := [][]Slot{
					{"Unknown Symbol", "Wizard"},
				}
				data, err := json.Marshal(lookupData)
				if err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(lookupDir, "lookup.json"), data, 0644)
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

			result := readLookupTableFromFile(tt.guildID)

			if tt.expectNil {
				if result != nil {
					t.Errorf("readLookupTableFromFile() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Errorf("readLookupTableFromFile() = nil, want non-nil")
				return
			}

			if result.GuildID != tt.guildID {
				t.Errorf("readLookupTableFromFile().GuildID = %v, want %v", result.GuildID, tt.guildID)
			}

			if len(result.Reels) != tt.reelCount {
				t.Errorf("readLookupTableFromFile() returned %d reels, want %d", len(result.Reels), tt.reelCount)
			}
		})
	}
}

func TestGetLookupTable(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "slots_lookup_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create the nested directory structure
	lookupDir := filepath.Join(tempDir, "slots", "lookuptable")
	symbolsDir := filepath.Join(tempDir, "slots", "symbols")
	err = os.MkdirAll(lookupDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
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

	// Create symbols file
	symbols := []Symbol{
		{Name: "Test Symbol 1", Emoji: "🎰"},
		{Name: "Test Symbol 2", Emoji: "🎲"},
	}
	symbolsData, err := json.Marshal(symbols)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(symbolsDir, "symbols.json"), symbolsData, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create lookup table file
	lookupData := [][]Slot{
		{"Test Symbol 1", "Test Symbol 2"},
		{"Test Symbol 2", "Test Symbol 1"},
	}
	lookupFileData, err := json.Marshal(lookupData)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(lookupDir, "lookup.json"), lookupFileData, 0644)
	if err != nil {
		t.Fatal(err)
	}

	guildID := "test-guild-123"
	result := GetLookupTable(guildID)

	if result == nil {
		t.Errorf("GetLookupTable() = nil, want non-nil")
		return
	}

	if result.GuildID != guildID {
		t.Errorf("GetLookupTable().GuildID = %v, want %v", result.GuildID, guildID)
	}

	if len(result.Reels) != 2 {
		t.Errorf("GetLookupTable() returned %d reels, want 2", len(result.Reels))
	}
}

func TestNewLookupTable(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "slots_lookup_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create the nested directory structure
	lookupDir := filepath.Join(tempDir, "slots", "lookuptable")
	symbolsDir := filepath.Join(tempDir, "slots", "symbols")
	err = os.MkdirAll(lookupDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
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

	// Test with missing files (should return nil)
	guildID := "test-guild-123"
	result := newLookupTable(guildID)
	if result != nil {
		t.Errorf("newLookupTable() with missing files = %v, want nil", result)
	}

	// Create symbols file
	symbols := []Symbol{
		{Name: "New Symbol 1", Emoji: "⭐"},
		{Name: "New Symbol 2", Emoji: "💎"},
	}
	symbolsData, err := json.Marshal(symbols)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(symbolsDir, "symbols.json"), symbolsData, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create lookup table file
	lookupData := [][]Slot{
		{"New Symbol 1", "New Symbol 2"},
	}
	lookupFileData, err := json.Marshal(lookupData)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(lookupDir, "lookup.json"), lookupFileData, 0644)
	if err != nil {
		t.Fatal(err)
	}

	result = newLookupTable(guildID)

	if result == nil {
		t.Errorf("newLookupTable() = nil, want non-nil")
		return
	}

	if result.GuildID != guildID {
		t.Errorf("newLookupTable().GuildID = %v, want %v", result.GuildID, guildID)
	}

	if len(result.Reels) != 1 {
		t.Errorf("newLookupTable() returned %d reels, want 1", len(result.Reels))
	}

	if len(result.Reels[0]) != 2 {
		t.Errorf("newLookupTable() first reel has %d symbols, want 2", len(result.Reels[0]))
	}
}

func TestLookupTable_GetCurrentSpin(t *testing.T) {
	// Create a test lookup table
	lookupTable := &LookupTable{
		GuildID: "test-guild",
		Reels: []Reel{
			{
				{Name: "Symbol1", Emoji: "🎰"},
				{Name: "Symbol2", Emoji: "🎲"},
				{Name: "Symbol3", Emoji: "⭐"},
			},
			{
				{Name: "SymbolA", Emoji: "💎"},
				{Name: "SymbolB", Emoji: "💰"},
			},
		},
	}

	// Run multiple times to test randomness
	for i := 0; i < 10; i++ {
		indices, spin := lookupTable.GetCurrentSpin()

		// Check that we got indices for each reel
		if len(indices) != 2 {
			t.Errorf("GetCurrentSpin() returned %d indices, want 2", len(indices))
		}

		// Check that we got symbols for each reel
		if len(spin) != 2 {
			t.Errorf("GetCurrentSpin() returned %d symbols, want 2", len(spin))
		}

		// Check that indices are within bounds
		if indices[0] < 0 || indices[0] >= len(lookupTable.Reels[0]) {
			t.Errorf("GetCurrentSpin() index[0] = %d, want 0 <= index < %d", indices[0], len(lookupTable.Reels[0]))
		}
		if indices[1] < 0 || indices[1] >= len(lookupTable.Reels[1]) {
			t.Errorf("GetCurrentSpin() index[1] = %d, want 0 <= index < %d", indices[1], len(lookupTable.Reels[1]))
		}

		// Check that the symbols match the indices
		if spin[0] != lookupTable.Reels[0][indices[0]] {
			t.Errorf("GetCurrentSpin() spin[0] does not match reel[0][%d]", indices[0])
		}
		if spin[1] != lookupTable.Reels[1][indices[1]] {
			t.Errorf("GetCurrentSpin() spin[1] does not match reel[1][%d]", indices[1])
		}
	}
}

func TestLookupTable_GetPreviousIndex(t *testing.T) {
	// Create a test reel with some repeated symbols
	reel := []Symbol{
		{Name: "A", Emoji: "🎰"},
		{Name: "A", Emoji: "🎰"}, // Duplicate symbol
		{Name: "B", Emoji: "🎲"},
		{Name: "C", Emoji: "⭐"},
		{Name: "C", Emoji: "⭐"}, // Duplicate symbol
		{Name: "A", Emoji: "🎰"},
	}

	lookupTable := &LookupTable{
		GuildID: "test-guild",
		Reels:   []Reel{reel},
	}

	tests := []struct {
		name         string
		currentIndex int
		expectedName string
	}{
		{
			name:         "from index 0",
			currentIndex: 0,
			expectedName: "C", // Should wrap to last different symbol
		},
		{
			name:         "from index 2",
			currentIndex: 2,
			expectedName: "A", // Previous different symbol
		},
		{
			name:         "from index 3",
			currentIndex: 3,
			expectedName: "B", // Previous different symbol
		},
		{
			name:         "from duplicate symbol",
			currentIndex: 1,
			expectedName: "C", // Should skip duplicate and go to different symbol
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			previousIndex := lookupTable.GetPreviousIndex(reel, tt.currentIndex)

			if previousIndex < 0 || previousIndex >= len(reel) {
				t.Errorf("GetPreviousIndex() = %d, want 0 <= index < %d", previousIndex, len(reel))
				return
			}

			previousSymbol := reel[previousIndex]
			if previousSymbol.Name != tt.expectedName {
				t.Errorf("GetPreviousIndex() returned symbol %v, want %v", previousSymbol.Name, tt.expectedName)
			}

			// Ensure the returned symbol is different from current
			currentSymbol := reel[tt.currentIndex]
			if previousSymbol.Name == currentSymbol.Name {
				t.Errorf("GetPreviousIndex() returned same symbol as current: %v", currentSymbol.Name)
			}
		})
	}
}

func TestLookupTable_GetNextIndex(t *testing.T) {
	// Create a test reel with some repeated symbols
	reel := []Symbol{
		{Name: "A", Emoji: "🎰"},
		{Name: "B", Emoji: "🎲"},
		{Name: "B", Emoji: "🎲"}, // Duplicate symbol
		{Name: "C", Emoji: "⭐"},
		{Name: "A", Emoji: "🎰"},
		{Name: "A", Emoji: "🎰"}, // Duplicate symbol
	}

	lookupTable := &LookupTable{
		GuildID: "test-guild",
		Reels:   []Reel{reel},
	}

	tests := []struct {
		name         string
		currentIndex int
		expectedName string
	}{
		{
			name:         "from index 0",
			currentIndex: 0,
			expectedName: "B", // Next different symbol
		},
		{
			name:         "from index 1",
			currentIndex: 1,
			expectedName: "C", // Should skip duplicate B
		},
		{
			name:         "from last index",
			currentIndex: 5,
			expectedName: "B", // Should wrap to first different symbol
		},
		{
			name:         "from index 3",
			currentIndex: 3,
			expectedName: "A", // Next different symbol
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextIndex := lookupTable.GetNextIndex(reel, tt.currentIndex)

			if nextIndex < 0 || nextIndex >= len(reel) {
				t.Errorf("GetNextIndex() = %d, want 0 <= index < %d", nextIndex, len(reel))
				return
			}

			nextSymbol := reel[nextIndex]
			if nextSymbol.Name != tt.expectedName {
				t.Errorf("GetNextIndex() returned symbol %v, want %v", nextSymbol.Name, tt.expectedName)
			}

			// Ensure the returned symbol is different from current
			currentSymbol := reel[tt.currentIndex]
			if nextSymbol.Name == currentSymbol.Name {
				t.Errorf("GetNextIndex() returned same symbol as current: %v", currentSymbol.Name)
			}
		})
	}
}

func TestLookupTable_GetPreviousSpin(t *testing.T) {
	// Create a test lookup table
	lookupTable := &LookupTable{
		GuildID: "test-guild",
		Reels: []Reel{
			{
				{Name: "A", Emoji: "🎰"},
				{Name: "B", Emoji: "🎲"},
				{Name: "C", Emoji: "⭐"},
			},
			{
				{Name: "X", Emoji: "💎"},
				{Name: "Y", Emoji: "💰"},
			},
		},
	}

	currentIndices := []int{1, 0} // B, X
	previousIndices, previousSpin := lookupTable.GetPreviousSpin(currentIndices)

	// Check that we got indices for each reel
	if len(previousIndices) != 2 {
		t.Errorf("GetPreviousSpin() returned %d indices, want 2", len(previousIndices))
	}

	// Check that we got symbols for each reel
	if len(previousSpin) != 2 {
		t.Errorf("GetPreviousSpin() returned %d symbols, want 2", len(previousSpin))
	}

	// Check that indices are within bounds
	for i, index := range previousIndices {
		if index < 0 || index >= len(lookupTable.Reels[i]) {
			t.Errorf("GetPreviousSpin() index[%d] = %d, want 0 <= index < %d", i, index, len(lookupTable.Reels[i]))
		}
	}

	// Check that the symbols match the indices
	for i, symbol := range previousSpin {
		if symbol != lookupTable.Reels[i][previousIndices[i]] {
			t.Errorf("GetPreviousSpin() spin[%d] does not match reel[%d][%d]", i, i, previousIndices[i])
		}
	}

	// Check that previous symbols are different from current symbols
	currentSpin := []Symbol{
		lookupTable.Reels[0][currentIndices[0]],
		lookupTable.Reels[1][currentIndices[1]],
	}
	for i, previousSymbol := range previousSpin {
		if previousSymbol.Name == currentSpin[i].Name {
			t.Errorf("GetPreviousSpin() returned same symbol as current at index %d: %v", i, currentSpin[i].Name)
		}
	}
}

func TestLookupTable_GetNextSpin(t *testing.T) {
	// Create a test lookup table
	lookupTable := &LookupTable{
		GuildID: "test-guild",
		Reels: []Reel{
			{
				{Name: "A", Emoji: "🎰"},
				{Name: "B", Emoji: "🎲"},
				{Name: "C", Emoji: "⭐"},
			},
			{
				{Name: "X", Emoji: "💎"},
				{Name: "Y", Emoji: "💰"},
			},
		},
	}

	currentIndices := []int{0, 1} // A, Y
	nextIndices, nextSpin := lookupTable.GetNextSpin(currentIndices)

	// Check that we got indices for each reel
	if len(nextIndices) != 2 {
		t.Errorf("GetNextSpin() returned %d indices, want 2", len(nextIndices))
	}

	// Check that we got symbols for each reel
	if len(nextSpin) != 2 {
		t.Errorf("GetNextSpin() returned %d symbols, want 2", len(nextSpin))
	}

	// Check that indices are within bounds
	for i, index := range nextIndices {
		if index < 0 || index >= len(lookupTable.Reels[i]) {
			t.Errorf("GetNextSpin() index[%d] = %d, want 0 <= index < %d", i, index, len(lookupTable.Reels[i]))
		}
	}

	// Check that the symbols match the indices
	for i, symbol := range nextSpin {
		if symbol != lookupTable.Reels[i][nextIndices[i]] {
			t.Errorf("GetNextSpin() spin[%d] does not match reel[%d][%d]", i, i, nextIndices[i])
		}
	}

	// Check that next symbols are different from current symbols
	currentSpin := []Symbol{
		lookupTable.Reels[0][currentIndices[0]],
		lookupTable.Reels[1][currentIndices[1]],
	}
	for i, nextSymbol := range nextSpin {
		if nextSymbol.Name == currentSpin[i].Name {
			t.Errorf("GetNextSpin() returned same symbol as current at index %d: %v", i, currentSpin[i].Name)
		}
	}
}

func TestLookupTable_Spin(t *testing.T) {
	// Create a test lookup table
	lookupTable := &LookupTable{
		GuildID: "test-guild",
		Reels: []Reel{
			{
				{Name: "A", Emoji: "🎰"},
				{Name: "B", Emoji: "🎲"},
				{Name: "C", Emoji: "⭐"},
			},
			{
				{Name: "X", Emoji: "💎"},
				{Name: "Y", Emoji: "💰"},
				{Name: "Z", Emoji: "🏆"},
			},
		},
	}

	// Test multiple spins
	for i := 0; i < 5; i++ {
		spin := lookupTable.Spin()

		if spin == nil {
			t.Errorf("Spin() returned nil")
			continue
		}

		// Check that we got the expected number of spins
		if len(spin.Spins) != NUM_SPINS {
			t.Errorf("Spin() returned %d spins, want %d", len(spin.Spins), NUM_SPINS)
		}

		// Check that win index is correct
		expectedWinIndex := NUM_SPINS - 2
		if spin.WinIndex != expectedWinIndex {
			t.Errorf("Spin() WinIndex = %d, want %d", spin.WinIndex, expectedWinIndex)
		}

		// Check that each spin has the correct number of symbols (one per reel)
		for j, singleSpin := range spin.Spins {
			if len(singleSpin) != len(lookupTable.Reels) {
				t.Errorf("Spin() spin[%d] has %d symbols, want %d", j, len(singleSpin), len(lookupTable.Reels))
			}
		}

		// Check that adjacent spins are different (animation effect)
		for j := 0; j < len(spin.Spins)-1; j++ {
			currentSpin := spin.Spins[j]
			nextSpin := spin.Spins[j+1]

			// At least one symbol should be different between adjacent spins
			hasDifference := false
			for k := 0; k < len(currentSpin); k++ {
				if currentSpin[k].Name != nextSpin[k].Name {
					hasDifference = true
					break
				}
			}

			if !hasDifference {
				t.Errorf("Spin() adjacent spins %d and %d are identical", j, j+1)
			}
		}
	}
}
