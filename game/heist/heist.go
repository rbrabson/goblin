package heist

import (
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/stats"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
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
	GuildID      string
	Organizer    *HeistMember
	Crew         []*HeistMember
	StartTime    time.Time
	State        HeistState
	interaction  *discordgo.InteractionCreate
	config       *Config
	mutex        sync.Mutex
	goodMessages []*HeistMessage
	badMessages  []*HeistMessage
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

	config := GetConfig(guildID)
	heist := &Heist{
		GuildID:      guildID,
		Organizer:    getHeistMember(guildID, memberID),
		Crew:         make([]*HeistMember, 0, 10),
		StartTime:    time.Now(),
		State:        Planning,
		config:       config,
		mutex:        sync.Mutex{},
		goodMessages: make([]*HeistMessage, 0, len(config.Theme.EscapedMessages)),
		badMessages:  make([]*HeistMessage, 0, len(config.Theme.ApprehendedMessages)+len(config.Theme.DiedMessages)),
	}

	err := heistChecks(heist, heist.Organizer)
	if err != nil {
		slog.Debug("heist checks failed", slog.String("guildID", guildID), slog.String("memberID", memberID), slog.Any("error", err))
		return nil, err
	}

	heist.Organizer.heist = heist
	heist.Crew = append(heist.Crew, heist.Organizer)
	currentHeists[guildID] = heist

	slog.Debug("create heist", slog.String("guildID", guildID), slog.String("memberID", memberID))

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
	slog.Debug("member joined heist", slog.String("guildID", member.GuildID), slog.String("memberID", member.MemberID))
	return nil
}

// Start runs the heist and returns the results of the heist.
func (h *Heist) Start() (*HeistResult, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if len(h.Crew) < 2 {
		h.State = Cancelled
		slog.Error("not enough members to start heist", slog.String("guildID", h.GuildID))
		return nil, ErrNotEnoughMembers{h.config.Theme}
	}

	h.State = InProgress
	target := getTarget(h.config.Targets, len(h.Crew))

	results := &HeistResult{
		AllResults:  make([]*HeistMemberResult, 0, len(h.Crew)),
		Escaped:     make([]*HeistMemberResult, 0, len(h.Crew)),
		Apprehended: make([]*HeistMemberResult, 0, len(h.Crew)),
		Dead:        make([]*HeistMemberResult, 0, len(h.Crew)),
		heist:       h,
		Target:      target,
	}

	successRate := calculateSuccessRate(h, target)
	for _, crewMember := range h.Crew {
		guildMember := crewMember.guildMember
		heistMember := getHeistMember(guildMember.GuildID, guildMember.MemberID)

		chance := rand.Float64()
		if chance <= successRate {
			goodResult := h.getGoodResult()
			bonus, msg := h.getBonusAmount(goodResult)

			result := &HeistMemberResult{
				Player:       heistMember,
				Status:       Free,
				Message:      msg,
				BonusCredits: bonus,
				heist:        h,
			}
			results.Escaped = append(results.Escaped, result)
			results.AllResults = append(results.AllResults, result)
		} else {
			badResult := h.getBadResult()

			result := &HeistMemberResult{
				Player:       heistMember,
				Status:       badResult.Result,
				Message:      badResult.Message,
				BonusCredits: 0,
				heist:        h,
			}
			if result.Status == Dead {
				results.Dead = append(results.Dead, result)
			} else {
				results.Apprehended = append(results.Apprehended, result)
			}
			results.AllResults = append(results.AllResults, result)
		}
	}

	// If at least one member escaped, then calculate the credits to distributed.
	if len(results.Escaped) > 0 {
		calculateCredits(results)
	}

	slog.Info("heist results",
		slog.String("guildID", h.GuildID),
		slog.Int("escaped", len(results.Escaped)),
		slog.Int("apprehended", len(results.Apprehended)),
		slog.Int("died", len(results.Dead)),
	)
	for _, result := range results.AllResults {
		slog.Info("heist member result",
			slog.String("member", result.Player.guildMember.Name),
			slog.Any("status", result.Status),
			slog.Int("payment", result.StolenCredits),
			slog.String("message", result.Message),
		)
	}

	return results, nil
}

// getGoodResult returns a random good result message, removing it from the list of available good messages
// to ensure that each message is only used once per heist.
func (h *Heist) getGoodResult() *HeistMessage {
	if len(h.goodMessages) == 0 {
		h.goodMessages = append(h.goodMessages, h.config.Theme.EscapedMessages...)
	}
	index := rand.IntN(len(h.goodMessages))
	message := h.goodMessages[index]

	h.goodMessages = append(h.goodMessages[:index], h.goodMessages[index+1:]...)

	return message
}

// getBadResult returns a random bad result message, removing it from the list of available bad messages
// to ensure that each message is only used once per heist.
func (h *Heist) getBadResult() *HeistMessage {
	if len(h.badMessages) == 0 {
		h.badMessages = append(h.badMessages, h.config.Theme.ApprehendedMessages...)
		h.badMessages = append(h.badMessages, h.config.Theme.DiedMessages...)
	}
	index := rand.IntN(len(h.badMessages))
	message := h.badMessages[index]

	h.badMessages = append(h.badMessages[:index], h.badMessages[index+1:]...)

	return message
}

// getBonusAmount calculates the bonus amount for a given good message, based on the heist's boost configuration.
// It returns the bonus amount and the updated message to reflect the new bonus amount. If there is no boost in
// effect, it simply returns the original bonus amount and message.
func (h *Heist) getBonusAmount(goodMessage *HeistMessage) (int, string) {
	if !h.config.BoostEnabled || h.config.BoostPercentage <= 0 {
		return goodMessage.BonusAmount, goodMessage.Message
	}

	msg := goodMessage.Message
	multiplier := 1.0 + (h.config.BoostPercentage / 100.0)
	bonus := int(float64(goodMessage.BonusAmount) * multiplier)

	// Update the message to reflect the new bonus amount
	strs := strings.Split(msg, "+")
	if len(strs) > 1 {
		strs2 := strings.Split(strings.TrimSpace(strs[1]), " ")
		if len(strs2) > 1 {
			p := message.NewPrinter(language.AmericanEnglish)
			msg = p.Sprintf("%s +%d %s", strings.TrimSpace(strs[0]), bonus, strings.TrimSpace(strs2[1]))
		}
	}
	return bonus, msg
}

// End ends the current heist, allowing for the cleanup of the heist.
// This is used when a heist is completed, and the results are being calculated.
func (h *Heist) End() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	heistCancelled := len(h.Crew) < 2
	if heistCancelled {
		h.State = Cancelled
		slog.Debug("heist cancelled", slog.String("guildID", h.GuildID))
	} else {
		h.State = Completed
		alertTimes[h.GuildID] = time.Now().Add(h.config.PoliceAlert)
		slog.Debug("heist ended", slog.String("guildID", h.GuildID))
	}

	heistLock.Lock()
	delete(currentHeists, h.GuildID)
	heistLock.Unlock()

	memberIDs := make([]string, 0, len(h.Crew))
	for _, member := range h.Crew {
		memberIDs = append(memberIDs, member.MemberID)
	}
	stats.UpdateGameStats(h.GuildID, "heist", memberIDs)
}

// Cancel cancels the current heist, removing it from the current heists.
func (h *Heist) Cancel() {
	heistLock.Lock()
	delete(currentHeists, h.GuildID)
	heistLock.Unlock()
}

// heistChecks returns an error, with appropriate message, if a heist cannot be started.
func heistChecks(h *Heist, member *HeistMember) error {
	if h.State != Planning {
		slog.Debug("heist already started", slog.String("guildID", h.GuildID), slog.String("state", string(h.State)))
		return ErrHeistAlreadyStarted
	}

	member.UpdateStatus()

	if slices.ContainsFunc(h.Crew, func(m *HeistMember) bool {
		return m.MemberID == member.MemberID
	}) {
		slog.Debug("member already joined heist", slog.String("guildID", h.GuildID), slog.String("memberID", member.MemberID))
		return ErrAlreadyJoinedHieist
	}

	account := bank.GetAccount(h.GuildID, member.MemberID)

	if account.CurrentBalance < h.config.HeistCost {
		return ErrNotEnoughCredits{h.config.HeistCost}
	}

	alertTime := alertTimes[h.GuildID]
	if alertTime.After(time.Now()) {
		remainingTime := time.Until(alertTime)
		return ErrPoliceOnAlert{h.config.Theme.Police, remainingTime}
	}

	if member.Status == Apprehended {
		remainingTime := member.RemainingJailTime()
		err := ErrInJail{h.config.Theme.Jail, h.config.Theme.Sentence, remainingTime, h.config.Theme.Bail, member.BailCost}
		return err
	}

	if member.Status == Dead {
		remainingTime := member.RemainingDeathTime()
		err := ErrDead{remainingTime}
		return err
	}

	return nil
}

// calculateSuccessRate returns the liklihood of a successful raid for each
// member of the heist crew.
func calculateSuccessRate(heist *Heist, target *Target) float64 {
	bonus := calculateBonusRate(heist, target)
	targetSuccess := target.Success / 100.0
	successChance := targetSuccess + bonus
	return successChance
}

// calculateBonusRate calculates the bonus amount to add to the success rate
// for a heist. The closer you are to the maximum crew size, the larger
// the bonus amount.
func calculateBonusRate(heist *Heist, target *Target) float64 {
	percent := float64(len(heist.Crew)) / float64(target.CrewSize)
	switch {
	case percent <= 0.2:
		return 0.0
	case percent <= 0.4:
		return 0.01
	case percent <= 0.6:
		return 0.03
	case percent <= 0.8:
		return 0.04
	default:
		return 0.05
	}
}

// calculateCredits determines the number of credits stolen by each surviving crew member.
func calculateCredits(results *HeistResult) {
	// Take 3/4 of the amount of the vault, and distribute it among those who survived (escaped or apprehended)
	numEscaped := len(results.Escaped)
	numApprehended := len(results.Apprehended)
	numSurvived := numEscaped + numApprehended
	stolenPerSurivor := int(math.Round(float64(results.Target.Vault) * 0.75 / float64(numSurvived)))

	config := results.heist.config
	if config.BoostEnabled && config.BoostPercentage > 0 {
		multiplier := 1.0 + (config.BoostPercentage / 100.0)
		stolenPerSurivor = int(float64(stolenPerSurivor) * multiplier)
	}
	totalStolen := min(numSurvived*stolenPerSurivor, results.Target.Vault)

	// Get a "base amount" of loot stolen. If you are apprehended, this is what you get. If you escaped you get 2x as much.
	baseStolen := totalStolen / (2*numEscaped + numApprehended)
	stolenPerEscaped := 2 * baseStolen
	stolenPerApprehended := baseStolen

	results.TotalStolen = 0
	for _, heistMemberResult := range results.Escaped {
		heistMemberResult.StolenCredits = stolenPerEscaped
		results.TotalStolen += heistMemberResult.StolenCredits
	}
	for _, heistMemberResult := range results.Apprehended {
		heistMemberResult.StolenCredits = stolenPerApprehended
		results.TotalStolen += heistMemberResult.StolenCredits
	}
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
