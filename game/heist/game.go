package heist

import (
	"time"

	"github.com/rbrabson/dgame/guild"
	log "github.com/sirupsen/logrus"
)

const (
	VAULT_UPDATE_TIME     = 1 * time.Minute // Update the vault once every minute
	VAULT_RECOVER_PERCENT = 1.04            // Percentage of valuts total that is recovered every update
)

var (
	heists = make(map[string]*Heist)
)

// startHeist creates a new heist with the given member as the heist planner
func startHeist(guild *guild.Guild, member *Member) error {
	log.Trace("--> heist.startHeist")
	defer log.Trace("<-- heist.startHeist")

	return nil
}

// joinHeist adds the given memmber to the heist crew
func joinHeist(heist *Heist, member *Member) error {
	log.Trace("--> heist.joinHeist")
	defer log.Trace("<-- heist.joinHeist")

	return nil
}

// getHeistResults returns the results of the heist
func getHeistResults(guild *guild.Guild) *HeistResult {
	log.Trace("--> heist.getHeistResults")
	defer log.Trace("<-- heist.getHeistResults")

	return nil
}

// resetHeist resets a hung heist
func resetHeist(guild *guild.Guild) error {
	log.Trace("--> heist.resetHeist")
	defer log.Trace("<-- heist.resetHeist")

	heist := heists[guild.ID]
	if heist == nil {
		log.WithFields(log.Fields{"guild": guild.ID}).Warn("heist not found")
		return ErrNoHeist
	}

	delete(heists, guild.ID)
	log.WithFields(log.Fields{"guild": guild.ID}).Info("heist reset")

	return nil
}

// getTarget returns the heist target given the number of crew members
func getTarget(heist *Heist, targets map[string]*Target) *Target {
	log.Trace("--> heist.getTarget")
	defer log.Trace("<-- heist.getTarget")

	crewSize := len(heist.CrewIDs)
	var target *Target
	for _, possibleTarget := range targets {
		if possibleTarget.CrewSize >= crewSize {
			if target == nil || target.CrewSize > possibleTarget.CrewSize {
				target = possibleTarget
			}
		}
	}

	log.WithFields(log.Fields{"guild": target.guildID, "target": target.ID}).Debug("heist target selected")
	return target
}
