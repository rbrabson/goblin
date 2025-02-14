package heist

import (
	"encoding/json"

	"github.com/rbrabson/dgame/guild"
	"github.com/rbrabson/dgame/mathex"
	log "github.com/sirupsen/logrus"
)

const (
	TARGET = "target"
)

var (
	targets = make(map[string]map[string]*Targets)
)

// Targets is the set of targets for a given theme
type Targets struct {
	ID      string
	Targets []Target
}

// Target is a target of a heist.
type Target struct {
	ID       string  `json:"_id" bson:"_id"`
	CrewSize int     `json:"crew" bson:"crew"`
	Success  float64 `json:"success" bson:"success"`
	Vault    int     `json:"vault" bson:"vault"`
	VaultMax int     `json:"vault_max" bson:"vault_max"`
	guildID  string  `json:"-" bson:"-"`
}

// NewTarget creates a new target for a heist
func NewTarget(id string, maxCrewSize int, success float64, vaultCurrent int, maxVault int) *Target {
	log.Debug("--> heist.NewTarget")
	defer log.Debug("<-- heist.NewTarget")

	target := Target{
		ID:       id,
		CrewSize: maxCrewSize,
		Success:  success,
		Vault:    vaultCurrent,
		VaultMax: maxVault,
	}
	return &target
}

// LoadTargets loads the targets that may be used in heists by the given guild
func LoadTargets(guild guild.Guild) {
	log.Debug("--> heist.LoadTargets")
	defer log.Debug("<-- heist.LoadTargets")

	themeTargets := make(map[string]*Targets)
	themes, _ := db.ListDocuments(guild.ID, TARGET)
	for _, theme := range themes {
		var t Targets
		db.Read(guild.ID, TARGET, theme, &t)
		themeTargets[t.ID] = &t
	}

	log.WithFields(log.Fields{"guild": guild.ID, "targets": themeTargets}).Trace("load targets")
	targets[guild.ID] = themeTargets
}

// GetTargets returns the specified list of targets for the server and theme
func GetTargets(guild guild.Guild, theme string) *Targets {
	log.Debug("--> heist.GetTargets")
	defer log.Debug("<-- heist.GetTargets")

	targetSet := targets[guild.ID][theme]
	if targetSet == nil {
		log.WithFields(log.Fields{"guild": guild.ID, "theme": theme}).Warning("targets not found")
		return nil
	}

	log.WithFields(log.Fields{"guild": guild.ID, "theme": theme, "targets": targetSet}).Trace("get targets")
	return targetSet
}

// Write writes the set of targets to the database. If they already exist, the are updated; otherwise, the set is created.
func (target *Target) Write() {
	log.Debug("--> heist.Target.Write")
	defer log.Debug("<-- heist.Target.Write")

	db.Write(target.guildID, TARGET, target.ID, targets)
	log.WithFields(log.Fields{"guild": target.guildID, "target": target.ID}).Trace("save target")
}

// vaultUpdater updates the vault balance for any target whose vault is not at the maximum value
func vaultUpdater() {
	// Update the vaults forever
	for {
		for _, guildTargets := range targets {
			for _, targets := range guildTargets {
				for _, target := range targets.Targets {
					newVaultAmount := int(float64(target.Vault) * VAULT_RECOVER_PERCENT)
					newVaultAmount = mathex.Min(newVaultAmount, target.VaultMax)
					if newVaultAmount != target.Vault {
						target.Write()
						log.WithFields(log.Fields{"guild": target.guildID, "target": target.ID, "Old": target.Vault, "New": newVaultAmount, "Max": target.VaultMax}).Debug("update vault")
					}
				}
			}
		}
	}
}

// String returns a string representation of the targets.
func (t *Targets) String() string {
	out, _ := json.Marshal(t)
	return string(out)
}

// String returns a string representation of the target.
func (t *Target) String() string {
	out, _ := json.Marshal(t)
	return string(out)
}
