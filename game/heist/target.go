package heist

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/rbrabson/goblin/discord"
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
	configFileName := filepath.Join(discord.DISCORD_CONFIG_DIR, "heist", "targets", HEIST_THEME+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		slog.Error("failed to read default targets",
			slog.String("guildID", guildID),
			slog.String("file", configFileName),
			slog.Any("error", err),
		)
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
	}
	for _, target := range targets {
		target.GuildID = guildID
		target.Theme = HEIST_THEME
		target.Vault = target.VaultMax
		target.IsAtMax = true
	}

	slog.Debug("create new targets",
		slog.String("guildID", guildID),
		slog.String("file", configFileName),
		slog.Int("targets", len(targets)),
	)

	return targets
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
	// Get all vaults not at the max value
	filter := bson.D{{Key: "is_at_max", Value: false}}

	// Update the vaults once a minute forever
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		for _, target := range getAllTargets(filter) {
			recoverAmount := int(float64(target.VaultMax) * VaultRecoverPercent)
			newVaultAmount := min(target.Vault+recoverAmount, target.VaultMax)
			slog.Debug("vault updater",
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
