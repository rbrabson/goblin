package slots

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPayout_String(t *testing.T) {
	tests := []struct {
		name     string
		payout   Payout
		contains []string // Check if result contains these substrings
	}{
		{
			name: "basic payout",
			payout: Payout{
				Win:    []Slot{"AQ", "GW", "BK"},
				Bet100: 240000,
				Bet200: 480000,
				Bet300: 1000000,
			},
			contains: []string{"Win: [AQ, GW, BK]", "Payouts: [100:", "200:", "300:"},
		},
		{
			name: "empty payout",
			payout: Payout{
				Win:    []Slot{},
				Bet100: 0,
				Bet200: 0,
				Bet300: 0,
			},
			contains: []string{"Win: []", "Payouts: [100:", "200:", "300:"},
		},
		{
			name: "single slot payout",
			payout: Payout{
				Win:    []Slot{"Archer"},
				Bet100: 100,
				Bet200: 200,
				Bet300: 300,
			},
			contains: []string{"Win: [Archer]", "Payouts: [100:", "200:", "300:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.payout.String()
			for _, substr := range tt.contains {
				if !stringContains(result, substr) {
					t.Errorf("Payout.String() = %v, want to contain %v", result, substr)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestPayoutAmount_String(t *testing.T) {
	tests := []struct {
		name         string
		payoutAmount PayoutAmount
		contains     []string // Check if result contains these substrings
	}{
		{
			name: "basic payout amount",
			payoutAmount: PayoutAmount{
				Win: []string{"AQ", "GW", "BK"},
				Bet: map[int]int{
					100: 240000,
					200: 480000,
					300: 1000000,
				},
			},
			contains: []string{"PayoutAmount{", "Win: [", "AQ", "GW", "BK", "Payouts: [", "]}"},
		},
		{
			name: "empty payout amount",
			payoutAmount: PayoutAmount{
				Win: []string{},
				Bet: map[int]int{},
			},
			contains: []string{"PayoutAmount{", "Win: []", "Payouts: [", "]}"},
		},
		{
			name: "single symbol payout amount",
			payoutAmount: PayoutAmount{
				Win: []string{"Archer"},
				Bet: map[int]int{
					100: 1000,
				},
			},
			contains: []string{"PayoutAmount{", "Win: [Archer]", "Payouts: [", "]}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.payoutAmount.String()
			for _, substr := range tt.contains {
				if !stringContains(result, substr) {
					t.Errorf("PayoutAmount.String() = %v, want to contain %v", result, substr)
				}
			}
		})
	}
}

// Remove the duplicate contains function since we renamed it to stringContains

func TestPayoutTable_String(t *testing.T) {
	objectID := primitive.NewObjectID()

	tests := []struct {
		name          string
		payoutTable   PayoutTable
		expectedStart string
	}{
		{
			name: "empty payout table",
			payoutTable: PayoutTable{
				ID:      objectID,
				GuildID: "12345",
				Payouts: []PayoutAmount{},
			},
			expectedStart: "PayoutTable{ID: " + objectID.Hex() + ", GuildID: 12345, Payouts: []}",
		},
		{
			name: "payout table with one payout",
			payoutTable: PayoutTable{
				ID:      objectID,
				GuildID: "12345",
				Payouts: []PayoutAmount{
					{
						Win: []string{"Archer"},
						Bet: map[int]int{100: 1000},
					},
				},
			},
			expectedStart: "PayoutTable{ID: " + objectID.Hex() + ", GuildID: 12345, Payouts: [",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.payoutTable.String()
			if tt.name == "empty payout table" {
				if result != tt.expectedStart {
					t.Errorf("PayoutTable.String() = %v, want %v", result, tt.expectedStart)
				}
			} else {
				if !stringContains(result, "PayoutTable{ID: "+objectID.Hex()) {
					t.Errorf("PayoutTable.String() does not contain expected ID")
				}
				if !stringContains(result, "GuildID: 12345") {
					t.Errorf("PayoutTable.String() does not contain expected GuildID")
				}
			}
		})
	}
}

func TestReadPayoutTableFromFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "slots_payout_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create the nested directory structure
	payoutDir := filepath.Join(tempDir, "slots", "payout")
	err = os.MkdirAll(payoutDir, 0755)
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
		expectPayout bool
		payoutCount  int
	}{
		{
			name: "valid payout table file",
			setupFile: func() error {
				payouts := []Payout{
					{
						Win:    []Slot{"AQ", "GW", "BK"},
						Bet100: 240000,
						Bet200: 480000,
						Bet300: 1000000,
					},
					{
						Win:    []Slot{"Archer", "Archer", "Archer"},
						Bet100: 5000,
						Bet200: 10000,
						Bet300: 15000,
					},
				}
				data, err := json.Marshal(payouts)
				if err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(payoutDir, "payout.json"), data, 0644)
			},
			guildID:      "12345",
			expectNil:    false,
			expectPayout: true,
			payoutCount:  2,
		},
		{
			name: "empty payout table file",
			setupFile: func() error {
				payouts := []Payout{}
				data, err := json.Marshal(payouts)
				if err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(payoutDir, "payout.json"), data, 0644)
			},
			guildID:      "12345",
			expectNil:    false,
			expectPayout: false,
			payoutCount:  0,
		},
		{
			name: "invalid JSON file",
			setupFile: func() error {
				return os.WriteFile(filepath.Join(payoutDir, "payout.json"), []byte("invalid json"), 0644)
			},
			guildID:   "12345",
			expectNil: true,
		},
		{
			name: "missing file",
			setupFile: func() error {
				// Remove the file if it exists
				os.Remove(filepath.Join(payoutDir, "payout.json"))
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

			result := readPayoutTableFromFile(tt.guildID)

			if tt.expectNil {
				if result != nil {
					t.Errorf("readPayoutTableFromFile() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Errorf("readPayoutTableFromFile() = nil, want non-nil")
				return
			}

			if result.GuildID != tt.guildID {
				t.Errorf("readPayoutTableFromFile().GuildID = %v, want %v", result.GuildID, tt.guildID)
			}

			if len(result.Payouts) != tt.payoutCount {
				t.Errorf("readPayoutTableFromFile() returned %d payouts, want %d", len(result.Payouts), tt.payoutCount)
			}

			// Verify payout structure conversion
			if tt.expectPayout && len(result.Payouts) > 0 {
				payout := result.Payouts[0]
				if len(payout.Win) == 0 {
					t.Errorf("readPayoutTableFromFile() first payout has no win symbols")
				}
				if len(payout.Bet) == 0 {
					t.Errorf("readPayoutTableFromFile() first payout has no bet amounts")
				}
				// Check that standard bet amounts are present
				expectedBets := []int{100, 200, 300}
				for _, bet := range expectedBets {
					if _, exists := payout.Bet[bet]; !exists {
						t.Errorf("readPayoutTableFromFile() missing bet amount %d", bet)
					}
				}
			}
		})
	}
}

func TestGetPayoutTable(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "slots_payout_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create the nested directory structure
	payoutDir := filepath.Join(tempDir, "slots", "payout")
	err = os.MkdirAll(payoutDir, 0755)
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

	// Create a valid payout file
	payouts := []Payout{
		{
			Win:    []Slot{"Test Symbol 1", "Test Symbol 2", "Test Symbol 3"},
			Bet100: 5000,
			Bet200: 10000,
			Bet300: 15000,
		},
	}
	data, err := json.Marshal(payouts)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(payoutDir, "payout.json"), data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		guildID string
	}{
		{
			name:    "get payout table for guild",
			guildID: "12345",
		},
		{
			name:    "get payout table for different guild",
			guildID: "67890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPayoutTable(tt.guildID)

			if result == nil {
				t.Errorf("GetPayoutTable() = nil, want non-nil")
				return
			}

			if result.GuildID != tt.guildID {
				t.Errorf("GetPayoutTable().GuildID = %v, want %v", result.GuildID, tt.guildID)
			}

			// Verify that payouts were loaded
			if len(result.Payouts) == 0 {
				t.Errorf("GetPayoutTable() returned empty payouts")
			}
		})
	}
}

func TestNewPayoutTable(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "slots_payout_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create the nested directory structure
	payoutDir := filepath.Join(tempDir, "slots", "payout")
	err = os.MkdirAll(payoutDir, 0755)
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
	result := newPayoutTable(guildID)
	if result != nil {
		t.Errorf("newPayoutTable() with missing files = %v, want nil", result)
	}

	// Create a valid payout file
	payouts := []Payout{
		{
			Win:    []Slot{"New Symbol 1", "New Symbol 2"},
			Bet100: 2000,
			Bet200: 4000,
			Bet300: 6000,
		},
	}
	data, err := json.Marshal(payouts)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(payoutDir, "payout.json"), data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	result = newPayoutTable(guildID)

	if result == nil {
		t.Errorf("newPayoutTable() = nil, want non-nil")
		return
	}

	if result.GuildID != guildID {
		t.Errorf("newPayoutTable().GuildID = %v, want %v", result.GuildID, guildID)
	}

	if len(result.Payouts) != 1 {
		t.Errorf("newPayoutTable() returned %d payouts, want 1", len(result.Payouts))
	}

	if len(result.Payouts[0].Win) != 2 {
		t.Errorf("newPayoutTable() first payout has %d symbols, want 2", len(result.Payouts[0].Win))
	}
}

func TestPayoutTable_GetPayoutAmount(t *testing.T) {
	// Create a test payout table
	payoutTable := &PayoutTable{
		GuildID: "test-guild",
		Payouts: []PayoutAmount{
			{
				Win: []string{"AQ", "GW", "BK"},
				Bet: map[int]int{
					100: 240000,
					200: 480000,
					300: 1000000,
				},
			},
			{
				Win: []string{"Archer", "Archer", "Archer"},
				Bet: map[int]int{
					100: 5000,
					200: 10000,
					300: 15000,
				},
			},
			{
				Win: []string{"Archer or Wizard", "Archer or Wizard", "Archer or Wizard"},
				Bet: map[int]int{
					100: 1000,
					200: 2000,
					300: 3000,
				},
			},
		},
	}

	tests := []struct {
		name     string
		bet      int
		spin     []Symbol
		expected int
	}{
		{
			name: "exact match - jackpot",
			bet:  100,
			spin: []Symbol{
				{Name: "AQ", Emoji: "🏹"},
				{Name: "GW", Emoji: "🛡️"},
				{Name: "BK", Emoji: "⚔️"},
			},
			expected: 240000,
		},
		{
			name: "exact match - triple archer",
			bet:  200,
			spin: []Symbol{
				{Name: "Archer", Emoji: "🏹"},
				{Name: "Archer", Emoji: "🏹"},
				{Name: "Archer", Emoji: "🏹"},
			},
			expected: 10000,
		},
		{
			name: "or match - archer or wizard combination",
			bet:  100,
			spin: []Symbol{
				{Name: "Archer", Emoji: "🏹"},
				{Name: "Wizard", Emoji: "🧙"},
				{Name: "Archer", Emoji: "🏹"},
			},
			expected: 1000,
		},
		{
			name: "no match",
			bet:  100,
			spin: []Symbol{
				{Name: "Barbarian", Emoji: "💪"},
				{Name: "Spell", Emoji: "✨"},
				{Name: "Dragon", Emoji: "🐉"},
			},
			expected: 0,
		},
		{
			name: "wrong bet amount",
			bet:  999,
			spin: []Symbol{
				{Name: "AQ", Emoji: "🏹"},
				{Name: "GW", Emoji: "🛡️"},
				{Name: "BK", Emoji: "⚔️"},
			},
			expected: 0,
		},
		{
			name: "mismatched spin length",
			bet:  100,
			spin: []Symbol{
				{Name: "AQ", Emoji: "🏹"},
				{Name: "GW", Emoji: "🛡️"},
			}, // Only 2 symbols instead of 3
			expected: 0,
		},
		{
			name: "higher bet amount",
			bet:  300,
			spin: []Symbol{
				{Name: "AQ", Emoji: "🏹"},
				{Name: "GW", Emoji: "🛡️"},
				{Name: "BK", Emoji: "⚔️"},
			},
			expected: 1000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := payoutTable.GetPayoutAmount(tt.bet, tt.spin)
			if result != tt.expected {
				t.Errorf("GetPayoutAmount() = %v, want %v", result, tt.expected)
			}
		})
	}
}
