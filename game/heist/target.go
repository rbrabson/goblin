package heist

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	VaultRecoverPercent = 0.04 // Percentage of valuts total that is recovered every update
)

// Target is a target of a heist.
type Target struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID  string             `json:"guild_id" bson:"guild_id"`
	Theme    string             `json:"theme" bson:"theme"`
	Name     string             `json:"target_id" bson:"target_id"`
	CrewSize int                `json:"crew" bson:"crew"`
	Success  float64            `json:"success" bson:"success"`
	Vault    int                `json:"vault" bson:"vault"`
	VaultMax int                `json:"vault_max" bson:"vault_max"`
	IsAtMax  bool               `json:"is_at_max" bson:"is_at_max"`
}

// GetTargets returns the list of targets for the server
func GetTargets(guildID string, theme string) []*Target {
	targets, _ := readTargets(guildID, theme)
	if targets == nil {
		targets = readTargetsFromFIle(guildID)
		for _, target := range targets {
			writeTarget(target)
		}

	}

	return targets
}

// StealFromValut removes the given amount from the vault of the target.
// If the amount is greater than the vault, the vault is set to 0.
func (t *Target) StealFromValut(amount int) {
	if amount <= 0 {
		slog.Debug("nothing stolen from the vault",
			slog.String("guildID", t.GuildID),
			slog.String("target", t.Name),
		)
		return
	}

	originalVaultAmount := t.Vault

	t.Vault -= amount
	if t.Vault < 0 {
		t.Vault = 0
	}
	t.IsAtMax = false

	writeTarget(t)

	slog.Debug("steal from vault",
		slog.String("guild", t.GuildID),
		slog.String("target", t.Name),
		slog.Int("amount", amount),
		slog.Int("original", originalVaultAmount),
		slog.Int("new", t.Vault),
	)
}

// getAllTargets returns all targets that match the filter.
func getAllTargets(filter bson.D) []*Target {
	allTargets, _ := readAllTargets(filter)

	return allTargets
}

// getTarget returns the target with the smallest maximum crew size that exceeds the number of
// crew members. If no target matches the criteria, then the target with the maximum crew size
// is used.
func getTarget(targets []*Target, crewSize int) *Target {
	var target *Target
	for _, possible := range targets {
		if possible.CrewSize >= crewSize {
			if target == nil || target.CrewSize > possible.CrewSize {
				target = possible
			}
		}
	}
	if target == nil {
		target = targets[len(targets)-1]
	}
	slog.Debug("heist target",
		slog.String("guildID", target.GuildID),
		slog.String("target", target.Name),
	)
	return target
}

// readTargetsFromFIle returns the default targets for a server.
// If the file is not found or cannot be decoded, the default targets are used.
func readTargetsFromFIle(guildID string) []*Target {
	configTheme := os.Getenv("DISCORD_DEFAULT_THEME")
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	configFileName := filepath.Join(configDir, "heist", "targets", configTheme+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		slog.Error("failed to read default targets",
			slog.String("guildID", guildID),
			slog.String("file", configFileName),
			slog.Any("error", err),
		)
		return getDefaultTargets(guildID)
	}

	var targets []*Target
	err = json.Unmarshal(bytes, &targets)
	if err != nil {
		slog.Error("failed to unmarshal default targets",
			slog.String("guildID", guildID),
			slog.String("file", configFileName),
			slog.String("targets", string(bytes)),
			slog.Any("error", err),
		)
		return getDefaultTargets(guildID)
	}
	for _, target := range targets {
		target.GuildID = guildID
		target.Theme = configTheme
		target.Vault = target.VaultMax
		target.IsAtMax = true
	}

	slog.Info("create new targets",
		slog.String("guildID", guildID),
		slog.String("file", configFileName),
		slog.Int("targets", len(targets)),
	)

	return targets
}

// getDefaultTargets returns the default targets for a server.
func getDefaultTargets(guildID string) []*Target {
	targets := []*Target{
		newTarget(guildID, "clash", "Goblin Forest", 2, 29.3, 16000),
		newTarget(guildID, "clash", "Goblin Outpost", 3, 20.65, 24000),
		newTarget(guildID, "clash", "Goblin Outpost", 3, 20.65, 24000),
		newTarget(guildID, "clash", "Rocky Fort", 5, 14.5, 42000),
		newTarget(guildID, "clash", "Goblin Gauntlet", 8, 9.5, 71000),
		newTarget(guildID, "clash", "Gobbotown", 11, 6.75, 101000),
		newTarget(guildID, "clash", "Fort Knobs", 14, 5.2, 133000),
		newTarget(guildID, "clash", "Bouncy Castle", 17, 4.25, 167000),
		newTarget(guildID, "clash", "Gobbo Campus", 21, 3.5, 213000),
		newTarget(guildID, "clash", "Walls Of Steel", 25, 2.91, 263000),
		newTarget(guildID, "clash", "Obsidian Tower", 29, 2.49, 314000),
		newTarget(guildID, "clash", "Queen's Gambit", 34, 2.15, 379000),
		newTarget(guildID, "clash", "Faulty Towers", 39, 1.86, 448000),
		newTarget(guildID, "clash", "Megamansion", 44, 1.64, 512000),
		newTarget(guildID, "clash", "P.e.k.k.a's Playhouse", 49, 1.46, 598000),
		newTarget(guildID, "clash", "Sherbet Towers", 55, 1.31, 688000),
	}

	return targets
}

// newTarget creates a new target for a heist
func newTarget(guildID string, theme string, name string, maxCrewSize int, success float64, maxVault int) *Target {
	target := Target{
		GuildID:  guildID,
		Theme:    theme,
		Name:     name,
		CrewSize: maxCrewSize,
		Success:  success,
		Vault:    maxVault,
		VaultMax: maxVault,
		IsAtMax:  true,
	}
	return &target
}

// ResetVaultsToMaximumValue resets all vaults in a guild to the maximum amount.
func ResetVaultsToMaximumValue(guildID string) {
	filter := bson.D{{Key: "guild_id", Value: guildID}}
	targets := getAllTargets(filter)
	for _, target := range targets {
		target.Vault = target.VaultMax
		target.IsAtMax = true
		writeTarget(target)
		slog.Info("reset vault to maximum",
			slog.String("guildID", guildID),
			slog.String("target", target.Name),
			slog.Int("vault", target.Vault),
		)
	}
}

// vaultUpdater updates the vault balance for any target whose vault is not at the maximum value
func vaultUpdater() {
	const timer = 1 * time.Minute

	filter := bson.D{{Key: "is_at_max", Value: false}}
	// Update the vaults forever
	for {
		time.Sleep(timer)
		for _, target := range getAllTargets(filter) {
			recoverAmount := int(float64(target.VaultMax) * VaultRecoverPercent)
			newVaultAmount := min(target.Vault+recoverAmount, target.VaultMax)
			slog.Info("vault updater",
				slog.String("guildID", target.GuildID),
				slog.String("target", target.Name),
				slog.Int("old", target.Vault),
				slog.Int("new", newVaultAmount),
				slog.Int("max", target.VaultMax),
			)
			target.Vault = newVaultAmount
			if target.Vault == target.VaultMax {
				target.IsAtMax = true
			}
			writeTarget(target)
		}
	}
}

// String returns a string representation of the Target.
func (t *Target) String() string {
	return fmt.Sprintf("Target{ID=%s, GuildID=%s, TargetID=%s, CrewSize=%d, Success=%.2f, Vault=%d, VaultMax=%d}",
		t.ID,
		t.GuildID,
		t.Name,
		t.CrewSize,
		t.Success,
		t.Vault,
		t.VaultMax,
	)
}
