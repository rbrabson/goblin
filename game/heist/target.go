package heist

import (
	"fmt"
	"time"

	"github.com/rbrabson/goblin/guild"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	VAULT_UPDATE_TIME     = 1 * time.Minute // Update the vault once every minute
	VAULT_RECOVER_PERCENT = 0.04            // Percentage of valuts total that is recovered every update
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
}

// GetTargets returns the list of targets for the server
func GetTargets(g *guild.Guild, theme string) []*Target {
	log.Trace("--> heist.GetTargets")
	defer log.Trace("<-- heist.GetTargets")

	targets, _ := readTargets(g, theme)
	if targets == nil {
		targets = getDefaultTargets(g)
		for _, target := range targets {
			writeTarget(target)
		}

	}

	log.WithFields(log.Fields{"guild": g.GuildID, "targets": len(targets)}).Trace("get targets")
	return targets
}

// StealFromValut removes the given amount from the vault of the target.
// If the amount is greater than the vault, the vault is set to 0.
func (t *Target) StealFromValut(amount int) {
	log.Trace("--> heist.Target.StealFromValut")
	defer log.Trace("<-- heist.Target.StealFromValut")

	if amount <= 0 {
		log.WithField("amount", amount).Debug("nothing stolen from the vault")
		return
	}

	originalVaultAmount := t.Vault

	t.Vault -= amount
	if t.Vault < 0 {
		t.Vault = 0
	}
	writeTarget(t)

	log.WithFields(log.Fields{"guild": t.GuildID, "target": t.Name, "amount": amount, "original": originalVaultAmount, "new": t.Vault}).Debug("steal from vault")
}

// getAllTargets returns all targets that match the filter.
func getAllTargets(filter bson.D) []*Target {
	log.Trace("--> heist.getAllTargets")
	defer log.Trace("<-- heist.getAllTargets")

	allTargets, _ := readAllTargets(filter)

	return allTargets
}

// getTarget returns the target with the smallest maximum crew size that exceeds the number of
// crew members. If no target matches the criteria, then the target with the maximum crew size
// is used.
func getTarget(targets []*Target, crewSize int) *Target {
	log.Trace("--> heist.getTarget")
	defer log.Trace("<-- heist.getTarget")

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
	log.WithField("Target", target.Name).Debug("Heist Target")
	return target
}

// getDefaultTargets returns the default targets for a server.
func getDefaultTargets(g *guild.Guild) []*Target {
	log.Debug("--> heist.getDefaultTargets")
	defer log.Debug("<-- heist.getDefaultTargets")

	targets := []*Target{
		newTarget(g, "clash", "Goblin Forest", 2, 29.3, 16000, 16000),
		newTarget(g, "clash", "Goblin Outpost", 3, 20.65, 24000, 24000),
		newTarget(g, "clash", "Rocky Fort", 5, 14.5, 42000, 42000),
		newTarget(g, "clash", "Goblin Gauntlet", 8, 9.5, 71000, 71000),
		newTarget(g, "clash", "Gobbotown", 11, 6.75, 101000, 101000),
		newTarget(g, "clash", "Fort Knobs", 14, 5.2, 133000, 133000),
		newTarget(g, "clash", "Bouncy Castle", 17, 4.25, 167000, 167000),
		newTarget(g, "clash", "Gobbo Campus", 21, 3.5, 213000, 213000),
		newTarget(g, "clash", "Walls Of Steel", 25, 2.91, 263000, 263000),
		newTarget(g, "clash", "Obsidian Tower", 29, 2.49, 314000, 314000),
		newTarget(g, "clash", "Queen's Gambit", 34, 2.15, 379000, 379000),
		newTarget(g, "clash", "Faulty Towers", 39, 1.86, 448000, 448000),
		newTarget(g, "clash", "Megamansion", 44, 1.64, 512000, 512000),
		newTarget(g, "clash", "P.e.k.k.a's Playhouse", 49, 1.46, 598000, 598000),
		newTarget(g, "clash", "Sherbet Towers", 55, 1.31, 688000, 688000),
	}

	return targets
}

// newTarget creates a new target for a heist
func newTarget(guild *guild.Guild, theme string, name string, maxCrewSize int, success float64, vaultCurrent int, maxVault int) *Target {
	log.Debug("--> heist.newTarget")
	defer log.Debug("<-- heist.newTarget")

	target := Target{
		GuildID:  guild.GuildID,
		Theme:    theme,
		Name:     name,
		CrewSize: maxCrewSize,
		Success:  success,
		Vault:    vaultCurrent,
		VaultMax: maxVault,
	}
	return &target
}

// vaultUpdater updates the vault balance for any target whose vault is not at the maximum value
func vaultUpdater() {
	const timer = time.Duration(1 * time.Minute)

	filter := bson.D{{Key: "$where", Value: "this.vault < this.vault_max"}}
	// Update the vaults forever
	for {
		time.Sleep(timer)
		log.WithFields(log.Fields{"timer": timer}).Trace("vault updater")
		for _, target := range getAllTargets(filter) {
			recoverAmount := int(float64(target.VaultMax) * VAULT_RECOVER_PERCENT)
			newVaultAmount := min(target.Vault+recoverAmount, target.VaultMax)
			log.WithFields(log.Fields{"guild": target.GuildID, "target": target.Name, "old": target.Vault, "new": newVaultAmount, "max": target.VaultMax}).Info("vault updater: update vault")
			target.Vault = newVaultAmount
			writeTarget(target)
		}
	}
}

// String returns a string representation of the Target.
func (target *Target) String() string {
	return fmt.Sprintf("Target{ID=%s, GuildID=%s, TargetID=%s, CrewSize=%d, Success=%.2f, Vault=%d, VaultMax=%d}",
		target.ID,
		target.GuildID,
		target.Name,
		target.CrewSize,
		target.Success,
		target.Vault,
		target.VaultMax,
	)
}
