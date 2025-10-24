package blackjack

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	bj "github.com/rbrabson/blackjack"
	"github.com/rbrabson/cards"
)

var (
	gamesLock sync.Mutex
	games     = make(map[string]*Game)
)

type GameState int

const (
	_ GameState = iota
	NotStarted
	WaitingForPlayers
	InProgress
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
	guildID       string
	game          *bj.Game
	config        *Config
	state         GameState
	gameStartTime time.Time
	turnChan      chan Action
	interaction   *discordgo.InteractionCreate
	symbols       *Symbols
	lock          sync.Mutex
}

// GetGame retrieves the blackjack game for the specified guild.
// If no game exists, a new one is created.
func GetGame(guildID string) *Game {
	gamesLock.Lock()
	defer gamesLock.Unlock()

	if game, exists := games[guildID]; exists {
		return game
	}
	game := newGame(guildID)
	games[guildID] = game
	return game
}

// newGame creates a new blackjack game for the specified guild.
func newGame(guildID string) *Game {
	config := GetConfig()
	game := &Game{
		guildID:  guildID,
		game:     bj.New(config.Decks),
		config:   config,
		state:    NotStarted,
		turnChan: make(chan Action),
		symbols:  GetSymbols(),
		lock:     sync.Mutex{},
	}

	return game
}

// AddPlayer adds a player to the blackjack game with a chip manager that uses their bank account.
// If the player already exists, no action is taken.
func (g *Game) AddPlayer(memberID string) error {
	if g.IsActive() {
		return ErrGameActive
	}
	if g.GetPlayer(memberID) != nil {
		return ErrPlayerAlreadyInGame
	}
	if len(g.game.Players()) >= g.config.MaxPlayers {
		return ErrGameFull
	}

	cm := NewChipManager(g.guildID, memberID)
	g.game.AddPlayer(memberID, 0, bj.WithChipManager(cm))
	player := g.GetPlayer(memberID)
	if err := player.PlaceBet(g.config.BetAmount); err != nil {
		g.game.RemovePlayer(memberID)
		return err
	}

	if g.NotStarted() {
		g.state = WaitingForPlayers
	}

	// If this is the first player, set the game start time to wait for additional players.
	if len(g.game.Players()) == 1 {
		g.gameStartTime = time.Now().Add(g.config.WaitForPlayers)
	}

	return nil
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
	// If the game is already active, do nothing.
	if g.IsActive() {
		return nil
	}

	err := g.game.StartNewRound()
	if err != nil {
		return err
	}
	g.state = InProgress
	return nil
}

// EndRound ends the current round of blackjack for the guild, removing all players from the game.
func (g *Game) EndRound() {
	for _, player := range g.game.Players() {
		g.game.RemovePlayer(player.Name())
	}
	for len(g.turnChan) > 0 {
		<-g.turnChan
	}
	g.interaction = nil
	g.state = WaitingForPlayers
}

// NotStarted returns whether the blackjack game has not yet started.
func (g *Game) NotStarted() bool {
	return g.state == NotStarted
}

// IsActive returns whether the blackjack game is currently active.
func (g *Game) IsActive() bool {
	return g.state == InProgress
}

// IsWaitingForPlayers returns whether the blackjack game is waiting for players to join.
func (g *Game) IsWaitingForPlayers() bool {
	return g.state == WaitingForPlayers
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
	return g.game.DealInitialCards()
}

// Dealer returns the dealer of the blackjack game.
func (g *Game) Dealer() *bj.Dealer {
	return g.game.Dealer()
}

// PlayerHit processes a hit action for the specified player.
func (g *Game) PlayerHit(playerName string) error {
	return g.game.PlayerHit(playerName)
}

// PlayerStand processes a stand action for the specified player.
func (g *Game) PlayerStand(playerName string) error {
	return g.game.PlayerStand(playerName)
}

// PlayerDoubleDownHit processes a double down hit action for the specified player.
func (g *Game) PlayerDoubleDownHit(playerName string) error {
	return g.game.PlayerDoubleDownHit(playerName)
}

// PlayerSplit processes a split action for the specified player.
func (g *Game) PlayerSplit(playerName string) error {
	return g.game.PlayerSplit(playerName)
}

// PlayerSurrender processes a surrender action for the specified player.
func (g *Game) PlayerSurrender(playerName string) error {
	return g.game.PlayerSurrender(playerName)
}

// DealerPlay processes the dealer's play according to blackjack rules.
func (g *Game) DealerPlay() error {
	return g.game.DealerPlay()
}

// PayoutResults pays out the results of the blackjack game.
func (g *Game) PayoutResults() {
	g.game.PayoutResults()
}

// EvaluateHand evaluates the result of a specific hand for a player.
func (g *Game) EvaluateHand(player *bj.Player) bj.GameResult {
	return g.game.EvaluateHand(player)
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
