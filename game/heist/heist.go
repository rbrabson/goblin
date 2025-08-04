package heist

import (
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"slices"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/bank"
)

var (
	alertTimes    = make(map[string]time.Time)
	currentHeists = make(map[string]*Heist)
	heistLock     = sync.Mutex{}
)

type HeistState string

const (
	Planning   HeistState = "Planning"
	InProgress HeistState = "In Progress"
	Cancelled  HeistState = "Cancelled"
	Completed  HeistState = "Completed"
)

// Heist is a heist that is being planned, is in progress, or has completed
type Heist struct {
	GuildID     string
	Organizer   *HeistMember
	Crew        []*HeistMember
	StartTime   time.Time
	State       HeistState
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
	Status        MemberStatus
	Message       string
	StolenCredits int
	BonusCredits  int
	heist         *Heist
}

// GetHeist returns the current heist for the given guild ID. If there is no
// heist, it returns nil.
func GetHeist(guildID string) *Heist {
	heistLock.Lock()
	defer heistLock.Unlock()
	return currentHeists[guildID]
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
		State:     Planning,
		config:    GetConfig(guildID),
		targets:   GetTargets(guildID, theme.Name),
		theme:     theme,
		mutex:     sync.Mutex{},
	}
	heist.mutex.Lock()
	defer heist.mutex.Unlock()

	err := heistChecks(heist, organizer)
	if err != nil {
		slog.Debug("heist checks failed",
			slog.String("guildID", guildID),
			slog.String("memberID", memberID),
			slog.Any("error", err),
		)
		return nil, err
	}

	organizer.heist = heist
	heist.Crew = append(heist.Crew, organizer)
	currentHeists[guildID] = heist

	slog.Debug("create heist",
		slog.String("guildID", guildID),
		slog.String("memberID", memberID),
	)

	return heist, nil
}

// AddCrewMember adds a crew member to the heist
func (h *Heist) AddCrewMember(member *HeistMember) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	err := heistChecks(h, member)
	if err != nil {
		return err
	}

	member.heist = h
	h.Crew = append(h.Crew, member)
	slog.Debug("member joined heist",
		slog.String("guildID", member.GuildID),
		slog.String("memberID", member.MemberID),
	)
	return nil
}

// Start runs the heist and returns the results of the heist.
func (h *Heist) Start() (*HeistResult, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.State = InProgress

	if len(h.Crew) < 2 {
		slog.Error("not enough members to start heist",
			slog.String("guildID", h.GuildID),
		)
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

	successRate := calculateSuccessRate(h, target)

	source := rand.NewPCG(rand.Uint64(), rand.Uint64())
	r := rand.New(source)
	for _, crewMember := range h.Crew {
		guildMember := crewMember.guildMember
		chance := r.IntN(100) + 1
		slog.Debug("heist results",
			slog.String("member", guildMember.Name),
			slog.Int("chance", chance),
			slog.Int("successRate", successRate),
		)
		if chance <= successRate {
			index := r.IntN(len(goodResults))
			goodResult := goodResults[index]
			updatedResults := make([]*HeistMessage, 0, len(goodResults))
			updatedResults = append(updatedResults, goodResults[:index]...)
			goodResults = append(updatedResults, goodResults[index+1:]...)
			if len(goodResults) == 0 {
				goodResults = append(goodResults, h.theme.EscapedMessages...)
			}

			heistMember := getHeistMember(guildMember.GuildID, guildMember.MemberID)
			result := &HeistMemberResult{
				Player:       heistMember,
				Status:       Free,
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
			}

			heistMember := getHeistMember(guildMember.GuildID, guildMember.MemberID)
			result := &HeistMemberResult{
				Player:       heistMember,
				Status:       badResult.Result,
				Message:      badResult.Message,
				heist:        h,
				BonusCredits: 0,
			}
			if result.Status == Dead {
				results.Dead = append(results.Dead, result)
				results.AllResults = append(results.AllResults, result)
			} else {
				results.Apprehended = append(results.Apprehended, result)
				results.AllResults = append(results.AllResults, result)
			}
		}
	}

	// If at least one member escaped, then calculate the credits to distributed.
	slog.Debug("heist results",
		slog.Int("escaped", len(results.Escaped)),
		slog.Int("apprehended", len(results.Apprehended)),
		slog.Int("died", len(results.Dead)),
	)
	if len(results.Escaped) > 0 {
		calculateCredits(results)
	}

	return results, nil
}

// End ends the current heist, allowing for the cleanup of the heist.
// This is used when a heist is completed, and the results are being calculated.
func (h *Heist) End() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.State = Completed

	heistLock.Lock()
	delete(currentHeists, h.GuildID)
	alertTimes[h.GuildID] = time.Now().Add(h.config.PoliceAlert)
	heistLock.Unlock()

	slog.Debug("heist ended",
		slog.String("guildID", h.GuildID),
	)
}

// Cancel cancels the current heist, allowing for the cleanup of the heist.
// This is used when a heist is started, but the heist cannot be run for some reason.
func (h *Heist) Cancel() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.State = Cancelled

	heistLock.Lock()
	delete(currentHeists, h.GuildID)
	heistLock.Unlock()

	slog.Debug("heist cancelled",
		slog.String("guildID", h.GuildID),
	)
}

// heistChecks returns an error, with appropriate message, if a heist cannot be started.
func heistChecks(h *Heist, member *HeistMember) error {
	if h.State != Planning {
		slog.Debug("heist already started or completed",
			slog.String("guildID", h.GuildID),
			slog.String("state", string(h.State)),
		)
		return ErrHeistAlreadyStarted
	}

	member.UpdateStatus()

	if slices.ContainsFunc(h.Crew, func(m *HeistMember) bool {
		return m.MemberID == member.MemberID
	}) {
		slog.Debug("member already joined heist",
			slog.String("guildID", h.GuildID),
			slog.String("memberID", member.MemberID),
		)
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

	if member.Status == Apprehended {
		remainingTime := member.RemainingJailTime()
		err := &ErrInJail{h.theme.Jail, h.theme.Sentence, remainingTime, h.theme.Bail, member.BailCost}
		return err
	}

	if member.Status == Dead {
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
	targetSuccess := int(math.Round(target.Success))
	successChance := targetSuccess + bonus
	slog.Debug("success rate",
		slog.Int("bunusRate", bonus),
		slog.Int("targetSuccess", targetSuccess),
		slog.Int("successChance", successChance),
	)
	return successChance
}

// calculateBonusRate calculates the bonus amount to add to the success rate
// for a heist. The closer you are to the maximum crew size, the larger
// the bonus amount.
func calculateBonusRate(heist *Heist, target *Target) int {
	percent := 100 * len(heist.Crew) / target.CrewSize
	slog.Debug("percentage for calculating success bonus",
		slog.Int("crewSize", len(heist.Crew)),
		slog.Int("targetCrewSize", target.CrewSize),
		slog.Int("percent", percent),
	)
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
	slog.Debug("looted",
		slog.String("target", results.Target.Name),
		slog.Int("vault", results.Target.Vault),
		slog.Int("survivors", numSurvived),
		slog.Int("base", baseStolen),
	)

	results.TotalStolen = 0
	for _, heistMemberResult := range results.Escaped {
		heistMemberResult.StolenCredits = 2 * baseStolen
		results.TotalStolen += heistMemberResult.StolenCredits
	}
	for _, heistMemberResult := range results.Apprehended {
		heistMemberResult.StolenCredits = baseStolen
		results.TotalStolen += heistMemberResult.StolenCredits
	}
	slog.Debug("total stolen",
		slog.String("guildID", results.Target.GuildID),
		slog.String("target", results.Target.Name),
		slog.Int("totalStolen", results.TotalStolen),
	)
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
