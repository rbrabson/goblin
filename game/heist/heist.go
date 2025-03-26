package heist

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/bank"
	log "github.com/sirupsen/logrus"
)

var (
	alertTimes    = make(map[string]time.Time)
	currentHeists = make(map[string]*Heist)
	heistLock     = sync.Mutex{}
)

// Heist is a heist that is being planned, is in progress, or has completed
type Heist struct {
	GuildID     string
	Organizer   *HeistMember
	Crew        []*HeistMember
	StartTime   time.Time
	targets     []*Target
	theme       *Theme
	interaction *discordgo.InteractionCreate
	config      *Config
	mutex       sync.Mutex
}

// HeistResult are the results of a heist
type HeistResult struct {
	AllResults  []*HeistMemberResult
	Escaped     []*HeistMemberResult
	Apprehended []*HeistMemberResult
	Dead        []*HeistMemberResult
	Target      *Target
	TotalStolen int
	heist       *Heist
}

// HeistMemberResult is the result for a single member of the heist
type HeistMemberResult struct {
	Player        *HeistMember
	Status        string
	Message       string
	StolenCredits int
	BonusCredits  int
	heist         *Heist
}

// NewHeist creates a new heist if one is not already underway.
func NewHeist(guildID string, memberID string) (*Heist, error) {
	heistLock.Lock()
	defer heistLock.Unlock()

	if currentHeists[guildID] != nil {
		return nil, ErrHeistInProgress
	}

	theme := GetTheme(guildID)
	organizer := getHeistMember(guildID, memberID)
	heist := &Heist{
		GuildID:   guildID,
		Organizer: organizer,
		Crew:      make([]*HeistMember, 0, 10),
		StartTime: time.Now(),
		config:    GetConfig(guildID),
		targets:   GetTargets(guildID, theme.Name),
		theme:     theme,
		mutex:     sync.Mutex{},
	}

	err := heistChecks(heist, organizer)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "organizer": memberID, "error": err}).Debug("heist checks failed")
		return nil, err
	}

	organizer.heist = heist
	heist.Crew = append(heist.Crew, organizer)
	currentHeists[guildID] = heist

	log.WithFields(log.Fields{"guild": guildID, "organizer": memberID}).Debug("create heist")

	return heist, nil
}

// addCrewMember adds a crew member to the heist
func (h *Heist) AddCrewMember(member *HeistMember) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	err := heistChecks(h, member)
	if err != nil {
		return err
	}

	member.heist = h
	h.Crew = append(h.Crew, member)
	log.WithFields(log.Fields{"guild": h.GuildID, "member": member.MemberID}).Debug("member joined heist")
	return nil
}

// Start runs the heist and returns the results of the heist.
func (h *Heist) Start() (*HeistResult, error) {
	heistLock.Lock()
	defer heistLock.Unlock()

	if len(h.Crew) < 2 {
		log.WithFields(log.Fields{"guild": h.GuildID}).Error("not enough members to start heist")
		return nil, ErrNotEnoughMembers{*h.theme}
	}

	target := getTarget(h.targets, len(h.Crew))

	results := &HeistResult{
		AllResults:  make([]*HeistMemberResult, 0, len(h.Crew)),
		Escaped:     make([]*HeistMemberResult, 0, len(h.Crew)),
		Apprehended: make([]*HeistMemberResult, 0, len(h.Crew)),
		Dead:        make([]*HeistMemberResult, 0, len(h.Crew)),
		heist:       h,
		Target:      target,
	}
	goodResults := make([]*HeistMessage, 0, len(h.theme.EscapedMessages))
	badResults := make([]*HeistMessage, 0, len(h.theme.ApprehendedMessages)+len(h.theme.DiedMessages))
	goodResults = append(goodResults, h.theme.EscapedMessages...)
	badResults = append(badResults, h.theme.ApprehendedMessages...)
	badResults = append(badResults, h.theme.DiedMessages...)
	log.WithFields(log.Fields{"guild": h.GuildID, "goodResults": len(goodResults), "badResults": len(badResults)}).Trace("set good and bad result messages")

	successRate := calculateSuccessRate(h, target)

	source := rand.NewPCG(rand.Uint64(), rand.Uint64())
	r := rand.New(source)
	for _, crewMember := range h.Crew {
		guildMember := crewMember.guildMember
		chance := r.IntN(100) + 1
		log.WithFields(log.Fields{"Player": guildMember.Name, "Chance": chance, "SuccessRate": successRate}).Debug("Heist Results")
		if chance <= successRate {
			index := r.IntN(len(goodResults))
			goodResult := goodResults[index]
			updatedResults := make([]*HeistMessage, 0, len(goodResults))
			updatedResults = append(updatedResults, goodResults[:index]...)
			goodResults = append(updatedResults, goodResults[index+1:]...)
			if len(goodResults) == 0 {
				goodResults := append(goodResults, h.theme.EscapedMessages...)
				log.WithFields(log.Fields{"guild": h.GuildID, "goodResults": len(goodResults), "badResults": len(badResults)}).Trace("reset good result messages")
			}

			heistMember := getHeistMember(guildMember.GuildID, guildMember.MemberID)
			result := &HeistMemberResult{
				Player:       heistMember,
				Status:       FREE,
				Message:      goodResult.Message,
				BonusCredits: goodResult.BonusAmount,
				heist:        h,
			}
			results.Escaped = append(results.Escaped, result)
			results.AllResults = append(results.AllResults, result)
		} else {
			index := r.IntN(len(badResults))
			badResult := badResults[index]
			updatedResults := make([]*HeistMessage, 0, len(badResults))
			updatedResults = append(updatedResults, badResults[:index]...)
			badResults = append(updatedResults, badResults[index+1:]...)
			if len(badResults) == 0 {
				badResults = append(badResults, h.theme.ApprehendedMessages...)
				badResults = append(badResults, h.theme.DiedMessages...)
				log.WithFields(log.Fields{"guild": h.GuildID, "goodResults": len(goodResults), "badResults": len(badResults)}).Trace("reset bad result messages")
			}

			heistMember := getHeistMember(guildMember.GuildID, guildMember.MemberID)
			result := &HeistMemberResult{
				Player:       heistMember,
				Status:       string(badResult.Result),
				Message:      badResult.Message,
				heist:        h,
				BonusCredits: 0,
			}
			if result.Status == DEAD {
				results.Dead = append(results.Dead, result)
				results.AllResults = append(results.AllResults, result)
			} else {
				results.Apprehended = append(results.Apprehended, result)
				results.AllResults = append(results.AllResults, result)
			}
		}
	}

	// If at least one member escaped, then calculate the credits to distributed.
	log.WithFields(log.Fields{"Escaped": len(results.Escaped), "Apprehended": len(results.Apprehended), "Died": len(results.Dead)}).Debug("heist results")
	if len(results.Escaped) > 0 {
		calculateCredits(results)
	}

	return results, nil
}

// End ends the current heist, allowing for the cleanup of the heist.
func (h *Heist) End() {
	heistLock.Lock()
	defer heistLock.Unlock()
	delete(currentHeists, h.GuildID)
	alertTimes[h.GuildID] = time.Now().Add(h.config.PoliceAlert)

	log.WithFields(log.Fields{"guild": h.GuildID}).Debug("heist ended")
}

// heistChecks returns an error, with appropriate message, if a heist cannot be started.
func heistChecks(h *Heist, member *HeistMember) error {
	member.UpdateStatus()

	if slices.ContainsFunc(h.Crew, func(m *HeistMember) bool {
		return m.MemberID == member.MemberID
	}) {
		log.WithFields(log.Fields{"guild": h.GuildID, "member": member.MemberID}).Debug("member already joined heist")
		return ErrAlreadyJoinedHieist
	}

	account := bank.GetAccount(h.GuildID, member.MemberID)

	if account.CurrentBalance < h.config.HeistCost {
		return &ErrNotEnoughCredits{h.config.HeistCost}
	}

	alertTime := alertTimes[h.GuildID]
	if alertTime.After(time.Now()) {
		remainingTime := time.Until(alertTime)
		return &ErrPoliceOnAlert{h.theme.Police, remainingTime}
	}

	if member.Status == APPREHENDED {
		remainingTime := member.RemainingJailTime()
		err := &ErrInJail{h.theme.Jail, h.theme.Sentence, remainingTime, h.theme.Bail, member.BailCost}
		return err
	}

	if member.Status == DEAD {
		remainingTime := member.RemainingDeathTime()
		err := &ErrDead{remainingTime}
		return err
	}

	return nil
}

// calculateSuccessRate returns the liklihood of a successful raid for each
// member of the heist crew.
func calculateSuccessRate(heist *Heist, target *Target) int {
	bonus := calculateBonusRate(heist, target)
	successChance := int(math.Round(target.Success)) + bonus
	log.WithFields(log.Fields{"BonusRate": bonus, "TargetSuccess": math.Round(target.Success), "SuccessChance": successChance}).Debug("Success Rate")
	return successChance
}

// calculateBonusRate calculates the bonus amount to add to the success rate
// for a heist. The closer you are to the maximum crew size, the larger
// the bonus amount.
func calculateBonusRate(heist *Heist, target *Target) int {
	percent := 100 * len(heist.Crew) / target.CrewSize
	log.WithField("percent", percent).Debug("percentage for calculating success bonus")
	if percent <= 20 {
		return 0
	}
	if percent <= 40 {
		return 1
	}
	if percent <= 60 {
		return 3
	}
	if percent <= 80 {
		return 4
	}
	return 5
}

// calculateCredits determines the number of credits stolen by each surviving crew member.
func calculateCredits(results *HeistResult) {
	// Take 3/4 of the amount of the vault, and distribute it among those who survived.
	numEscaped := len(results.Escaped)
	numApprehended := len(results.Apprehended)
	numSurvived := numEscaped + numApprehended
	stolenPerSurivor := int(math.Round(float64(results.Target.Vault) * 0.75 / float64(numSurvived)))
	totalStolen := numSurvived * stolenPerSurivor

	// Get a "base amount" of loot stolen. If you are apprehended, this is what you get. If you escaped you get 2x as much.
	baseStolen := totalStolen / (2*numEscaped + numApprehended)

	results.TotalStolen = 0
	// Caculate a "base amount". Those who escape get 2x those who don't. So Divide the
	log.WithFields(log.Fields{"Target": results.Target.Name, "Vault": results.Target.Vault, "Survivors": numSurvived, "Base Credits": baseStolen}).Debug("Looted")
	for _, heistMemberResult := range results.Escaped {
		heistMemberResult.StolenCredits = 2 * baseStolen
		results.TotalStolen += heistMemberResult.StolenCredits
	}
	for _, heistMemberResult := range results.Apprehended {
		heistMemberResult.StolenCredits = baseStolen
		results.TotalStolen += heistMemberResult.StolenCredits
	}
	log.WithFields(log.Fields{"Guild": results.Target.GuildID, "Target": results.Target.Name, "TotalStolen": results.TotalStolen}).Debug("total stolen")
}

// String returns a string representation of the Heist.
func (h *Heist) String() string {
	return fmt.Sprintf("Heist{GuildID: %s, Organizer: %s, Crew: %d, StartTime: %s}",
		h.GuildID,
		h.Organizer,
		len(h.Crew),
		h.StartTime,
	)
}

// String returns a string representation of the HeistMember.
func (hr *HeistResult) String() string {
	return fmt.Sprintf("HeistResult{Escaped: %d, Apprehended: %d, Dead: %d, Target: %s, TotalStolen: %d}",
		len(hr.Escaped),
		len(hr.Apprehended),
		len(hr.Dead),
		hr.Target,
		hr.TotalStolen,
	)
}

// String returns a string representation of the HeistMemberResult.
func (hmr *HeistMemberResult) String() string {
	return fmt.Sprintf("HeistMemberResult{Player: %s, Status: %s, Message: %s, StolenCredits: %d, BonusCredits: %d}",
		hmr.Player,
		hmr.Status,
		hmr.Message,
		hmr.StolenCredits,
		hmr.BonusCredits,
	)
}
