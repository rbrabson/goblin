package blackjack

import (
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	bj "github.com/rbrabson/blackjack"
	"github.com/rbrabson/cards"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/stats"
)

var (
	gamesLock sync.Mutex
	games     = make(map[string]*Game)
)

type GameState int

const (
	_ GameState = iota
	NotStarted
	Starting
	WaitingForPlayers
	DealingHands
)

type Action int

const (
	_ Action = iota
	Hit
	Stand
	DoubleDown
	Split
	Surrender
)

// Game represents a blackjack game for a specific guild.
type Game struct {
	guildID          string
	game             *bj.Game
	config           *Config
	state            GameState
	gameStartTime    time.Time
	turnChan         chan Action
	interaction      *discordgo.InteractionCreate
	message          *discordgo.Message
	symbols          Symbols
	joinButton       discordgo.Button
	hitButton        discordgo.Button
	standButton      discordgo.Button
	doubleDownButton discordgo.Button
	splitButton      discordgo.Button
	surrenderButton  discordgo.Button
	uid              string
	lock             sync.Mutex
}

// GetGame retrieves the blackjack game for the specified guild.
// If no game exists, a new one is created.
func GetGame(guildID string, uid string) *Game {
	gamesLock.Lock()
	defer gamesLock.Unlock()

	config := GetConfig(guildID)
	game := games[uid]
	if game == nil {
		slog.Warn("blackjack game not found", slog.String("guildID", guildID), slog.String("uid", uid))
		return nil
	}
	game.config = config
	return game
}

// StartGame starts a new blackjack game for the specified guild and member.
func StartGame(guildID string, memberID string) (*Game, error) {
	gamesLock.Lock()
	defer gamesLock.Unlock()

	uid := getUID(guildID, memberID)
	config := GetConfig(guildID)
	game := games[uid]
	if game == nil {
		game = newGame(guildID, uid, config.Decks)
		games[uid] = game
		slog.Debug("created new blackjack game", slog.String("guildID", guildID), slog.String("uid", uid))
	} else {
		slog.Debug("retrieved existing blackjack game", slog.String("guildID", guildID), slog.String("uid", uid))
	}

	if !game.NotStarted() {
		slog.Debug("blackjack game has already started", slog.String("guildID", game.guildID), slog.String("memberID", memberID))
		return nil, ErrGameActive
	}

	game.config = config

	if err := game.addPlayer(memberID); err != nil {
		return nil, err
	}
	game.state = Starting

	return game, nil
}

// newGame creates a new blackjack game for the specified guild.
func newGame(guildID string, uid string, numDecks int) *Game {
	game := &Game{
		guildID:  guildID,
		uid:      uid,
		game:     bj.New(numDecks),
		state:    NotStarted,
		turnChan: make(chan Action, 5),
		symbols:  GetSymbols(),
		lock:     sync.Mutex{},
	}
	createButtons(game)

	return game
}

// joinGame allows a player to join the blackjack game if it has not started yet.
func (g *Game) joinGame(memberID string) error {
	g.Lock()
	defer g.Unlock()

	if g.NotStarted() {
		return ErrGameNotStarted
	}
	if !g.IsWaitingForPlayers() {
		return ErrGameActive
	}

	return g.addPlayer(memberID)
}

// addPlayer adds a player to the blackjack game with a chip manager that uses their bank account.
// If the player already exists, no action is taken.
func (g *Game) addPlayer(memberID string) error {
	if g.GetPlayer(memberID) != nil {
		return ErrPlayerAlreadyInGame
	}
	if g.IsDealingHands() {
		return ErrGameActive
	}
	if len(g.game.Players()) >= g.config.MaxPlayers {
		return ErrGameFull
	}

	cm := NewChipManager(g, memberID)
	g.game.AddPlayer(memberID, bj.WithChipManager(cm))
	player := g.GetPlayer(memberID)
	if err := player.CurrentHand().PlaceBet(g.config.BetAmount); err != nil {
		g.game.RemovePlayer(memberID)
		return err
	}

	if g.IsStarting() {
		g.state = WaitingForPlayers
	}

	// If this is the first player, set the game start time to wait for additional players.
	if len(g.game.Players()) == 1 {
		g.gameStartTime = time.Now().Add(g.config.WaitForPlayers)
	}

	return nil
}

// clearPendingActions clears any pending player actions from the turn channel, ensuring that
// no stale actions are processed when a new round starts or when a player takes an action.
func (g *Game) clearPendingActions() {
	g.Lock()
	defer g.Unlock()

	for len(g.turnChan) > 0 {
		<-g.turnChan
	}
}

// SetState sets the current state of the blackjack game.
func (g *Game) SetState(state GameState) {
	g.state = state
}

// GetPlayer retrieves a player from the blackjack game by their member ID.
func (g *Game) GetPlayer(memberID string) *bj.Player {
	return g.game.GetPlayer(memberID)
}

// GetACtivePlayer retrieves the currently active player in the blackjack game.
func (g *Game) GetActivePlayer() *bj.Player {
	return g.game.GetActivePlayer()
}

// Players returns a slice of all players in the blackjack game.
func (g *Game) Players() []*bj.Player {
	return g.game.Players()
}

// StartNewRound starts a new round of blackjack in the game.
func (g *Game) StartNewRound() error {
	g.Lock()
	defer g.Unlock()

	// If the game is already active, do nothing.
	if g.IsStarting() {
		return nil
	}

	if err := g.game.StartNewRound(); err != nil {
		return err
	}

	g.state = Starting
	return nil
}

// EndRound ends the current round of blackjack for the guild, removing all players from the game.
func (g *Game) EndRound() {
	g.Lock()
	defer g.Unlock()

	// Update the member stats
	for _, player := range g.Players() {
		slog.Debug("updating member stats for player", slog.String("guildID", g.guildID), slog.String("playerName", player.Name()))
		member := GetMember(g.guildID, player.Name())
		member.RoundPlayed(g, player)
	}

	memberIDs := make([]string, 0, len(g.Players()))
	for _, player := range g.Players() {
		memberIDs = append(memberIDs, player.Name())
	}
	stats.UpdateGameStats(g.guildID, "blackjack", memberIDs)

	for _, player := range g.game.Players() {
		slog.Debug("removing player from blackjack game", slog.String("guildID", g.guildID), slog.String("playerName", player.Name()))
		g.game.RemovePlayer(player.Name())
	}

	for len(g.turnChan) > 0 {
		<-g.turnChan
	}
	slog.Debug("cleared pending player actions for new round", slog.String("guildID", g.guildID))

	g.interaction = nil
	g.message = nil
	g.SetState(NotStarted)

	gamesLock.Lock()
	defer gamesLock.Unlock()

	if g.config.SinglePlayerMode {
		slog.Debug("deleting single player game", slog.String("guildID", g.guildID), slog.String("uid", g.uid))
		destroyButtons(g)
		delete(games, g.uid)
	} else {
		slog.Debug("clearing multiplayer game state for new round", slog.String("guildID", g.guildID))
		g.Dealer().ClearHand()
	}

	if status == discord.STOPPING {
		newstatus := discord.STOPPED
		for _, game := range games {
			if !game.NotStarted() {
				newstatus = discord.STOPPING
				break
			}
		}
		status = newstatus
	}
}

// NotStarted returns whether the blackjack game has not yet started.
func (g *Game) NotStarted() bool {
	return g.state == NotStarted
}

// IsStarting returns whether the blackjack game is currently active.
func (g *Game) IsStarting() bool {
	return g.state == Starting
}

// IsWaitingForPlayers returns whether the blackjack game is waiting for players to join.
func (g *Game) IsWaitingForPlayers() bool {
	return g.state == WaitingForPlayers
}

// IsDealingHands returns whether the blackjack game is currently dealing initial hands to players.
func (g *Game) IsDealingHands() bool {
	return g.state == DealingHands
}

// SecondsBeforeStart returns the number of seconds remaining to wait for players
// before starting the game. If the wait time has elapsed, it returns 0.
func (g *Game) SecondsBeforeStart() int {
	waitTime := time.Until(g.gameStartTime)
	if waitTime > 0 {
		return int(waitTime.Seconds())
	}
	return 0
}

// DealInitialCards deals the initial cards to all players and the dealer.
func (g *Game) DealInitialCards() error {
	g.Lock()
	defer g.Unlock()

	g.state = DealingHands

	if err := g.game.DealInitialCards(); err != nil {
		return err
	}
	for _, player := range g.game.Players() {
		for _, hand := range player.Hands() {
			hand.SetBet(g.config.BetAmount)
		}
	}
	return nil
}

// Dealer returns the dealer of the blackjack game.
func (g *Game) Dealer() *bj.Dealer {
	return g.game.Dealer()
}

// PlayerHit processes a hit action for the specified player.
func (g *Game) PlayerHit(player *bj.Player) error {
	g.Lock()
	defer g.Unlock()

	if err := g.game.PlayerHit(player.Name()); err != nil {
		return err
	}

	hand := player.CurrentHand()
	if hand.IsBusted() {
		slog.Debug("player busted", slog.String("guildID", g.guildID), slog.String("playerName", player.Name()), slog.Any("hand", hand))
		hand.SetActive(false)
	} else if hand.Value() == 21 {
		slog.Debug("player hand reached 21", slog.String("guildID", g.guildID), slog.String("playerName", player.Name()), slog.Any("hand", hand))
		hand.SetActive(false)
	}

	return nil
}

// PlayerStand processes a stand action for the specified player.
func (g *Game) PlayerStand(player *bj.Player) error {
	g.Lock()
	defer g.Unlock()

	if err := g.game.PlayerStand(player.Name()); err != nil {
		return err
	}

	hand := player.CurrentHand()
	hand.SetActive(false)

	return nil

}

// PlayerDoubleDown processes a double down hit action for the specified player.
func (g *Game) PlayerDoubleDown(player *bj.Player) error {
	g.Lock()
	defer g.Unlock()

	if !player.CurrentHand().CanDoubleDown() {
		slog.Error("cannot double down", slog.String("guildID", g.guildID), slog.String("playerName", player.Name()))
		return ErrCannotDoubleDown
	}
	if err := player.CurrentHand().DoubleDown(); err != nil {
		slog.Error("error processing player double down", slog.String("guildID", g.guildID), slog.String("playerName", player.Name()), slog.Any("error", err))
		return err
	}
	if err := g.game.PlayerDoubleDownHit(player.Name()); err != nil {
		slog.Error("error processing player double down hit", slog.String("guildID", g.guildID), slog.String("playerName", player.Name()), slog.Any("error", err))
		return err
	}

	hand := player.CurrentHand()
	if hand.IsBusted() {
		slog.Debug("player busted after double down", slog.String("guildID", g.guildID), slog.String("playerName", player.Name()), slog.Any("hand", hand))
	} else if hand.Value() == 21 {
		slog.Debug("player hand reached 21 after double down", slog.String("guildID", g.guildID), slog.String("playerName", player.Name()), slog.Any("hand", hand))
	}
	hand.SetActive(false)

	return nil
}

// PlayerSplit processes a split action for the specified player.
func (g *Game) PlayerSplit(player *bj.Player) error {
	g.Lock()
	defer g.Unlock()

	if !player.CurrentHand().CanSplit() {
		slog.Error("cannot split", slog.String("guildID", g.guildID), slog.String("playerName", player.Name()))
		return ErrCannotSplit
	}

	return g.game.PlayerSplit(player.Name())
}

// PlayerSurrender processes a surrender action for the specified player.
func (g *Game) PlayerSurrender(player *bj.Player) error {
	g.Lock()
	defer g.Unlock()

	if !player.CurrentHand().CanSurrender() {
		slog.Error("cannot surrender", slog.String("guildID", g.guildID), slog.String("playerName", player.Name()))
		return ErrCannotSurrender
	}

	if err := g.game.PlayerSurrender(player.Name()); err != nil {
		return err
	}

	hand := player.CurrentHand()
	hand.SetActive(false)

	return nil
}

// PlayerActionRequest processes a request from a player to take an action, ensuring that the player is active and the game is in progress.
func (g *Game) PlayerActionRequest(memberID string, action Action) error {
	g.Lock()
	defer g.Unlock()

	if !g.IsDealingHands() {
		return ErrGameNotStarted
	}

	player := g.GetPlayer(memberID)
	activePlayer := g.GetActivePlayer()
	if player == nil || player != activePlayer {
		return ErrNotActivePlayer
	}

	g.turnChan <- action
	return nil
}

// DealerPlay processes the dealer's play according to blackjack rules.
func (g *Game) DealerPlay() error {
	g.Lock()
	defer g.Unlock()

	if !g.hasNonbustedPlayers() {
		return ErrAllPlayersBusted
	}
	if err := g.game.DealerPlay(); err != nil {
		return err
	}

	return nil
}

// hasNonbustedPlayers checks if there are any players in the game who have not busted, surrendered, or gotten blackjack.
func (g *Game) hasNonbustedPlayers() bool {
	for _, player := range g.Players() {
		for _, hand := range player.Hands() {
			if !(hand.IsBusted() || hand.IsSurrendered() || hand.IsBlackjack()) {
				return true
			}
		}
	}
	return false
}

// PayoutResults pays out the results of the blackjack game.
func (g *Game) PayoutResults() {
	g.Lock()
	defer g.Unlock()

	g.game.PayoutResults()
}

// EvaluateHand evaluates the result of a specific hand for a player.
func (g *Game) EvaluateHand(hand *bj.Hand) bj.GameResult {
	return g.game.EvaluateHand(hand)
}

// Round returns the current round number of the blackjack game.
func (g *Game) Round() int {
	return g.game.Round()
}

// Lock locks the game's mutex.
func (g *Game) Lock() {
	g.lock.Lock()
}

// Unlock unlocks the game's mutex.
func (g *Game) Unlock() {
	g.lock.Unlock()
}

// handValue calculates the value of a blackjack hand.
func handValue(hand *bj.Hand, hidden bool) int {
	visibleValue := 0
	aces := 0
	for idx, card := range hand.Cards() {
		if hidden && idx == 0 {
			continue
		}

		rank := card.Rank
		switch rank {
		case cards.Jack, cards.Queen, cards.King:
			visibleValue += 10
		case cards.Ace:
			aces++
			visibleValue += 11
		default:
			visibleValue += int(rank)
		}
	}

	// Adjust for aces
	for aces > 0 && visibleValue > 21 {
		visibleValue -= 10
		aces--
	}

	return visibleValue
}

// createButtons creates and registers the action buttons for the blackjack game.
func createButtons(game *Game) {
	game.joinButton = discordgo.Button{
		Label:    "Join Game",
		Style:    discordgo.SuccessButton,
		CustomID: "blackjack_join" + ":" + game.uid,
	}
	bot.AddComponentHandler(game.joinButton.CustomID, blackjackJoin)

	game.hitButton = discordgo.Button{
		Label:    "Hit",
		Style:    discordgo.PrimaryButton,
		CustomID: "blackjack_hit" + ":" + game.uid,
	}
	bot.AddComponentHandler(game.hitButton.CustomID, blackjackHit)

	game.standButton = discordgo.Button{
		Label:    "Stand",
		Style:    discordgo.PrimaryButton,
		CustomID: "blackjack_stand" + ":" + game.uid,
	}
	bot.AddComponentHandler(game.standButton.CustomID, blackjackStand)

	game.doubleDownButton = discordgo.Button{
		Label:    "Double Down",
		Style:    discordgo.PrimaryButton,
		CustomID: "blackjack_double_down" + ":" + game.uid,
	}
	bot.AddComponentHandler(game.doubleDownButton.CustomID, blackjackDoubleDown)

	game.splitButton = discordgo.Button{
		Label:    "Split",
		Style:    discordgo.PrimaryButton,
		CustomID: "blackjack_split" + ":" + game.uid,
	}
	bot.AddComponentHandler(game.splitButton.CustomID, blackjackSplit)

	game.surrenderButton = discordgo.Button{
		Label:    "Surrender",
		Style:    discordgo.DangerButton,
		CustomID: "blackjack_surrender" + ":" + game.uid,
	}
	bot.AddComponentHandler(game.surrenderButton.CustomID, blackjackSurrender)
}

// destroyButtons deregisters the action buttons for the blackjack game.
func destroyButtons(game *Game) {
	bot.RemoveComponentHandler(game.joinButton.CustomID)
	bot.RemoveComponentHandler(game.hitButton.CustomID)
	bot.RemoveComponentHandler(game.standButton.CustomID)
	bot.RemoveComponentHandler(game.doubleDownButton.CustomID)
	bot.RemoveComponentHandler(game.splitButton.CustomID)
	bot.RemoveComponentHandler(game.surrenderButton.CustomID)
}

// getUID generates the unique identifier for the blackjack game based on the guild and member IDs.
func getUID(guildID string, memberID string) string {
	config := GetConfig(guildID)
	if config.SinglePlayerMode {
		return guildID + "-" + memberID
	}
	return guildID
}

// getUIDFromInteraction extracts the unique identifier from a Discord interaction.
func getUIDFromInteraction(i *discordgo.InteractionCreate) string {
	customID := i.Interaction.MessageComponentData().CustomID
	vars := strings.Split(customID, ":")
	var uid string
	if len(vars) == 1 {
		uid = vars[0]
	} else {
		uid = vars[1]
	}
	return uid
}
