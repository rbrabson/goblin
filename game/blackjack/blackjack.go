package blackjack

import (
	"sync"
	"time"

	bj "github.com/rbrabson/blackjack"
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

// Game represents a blackjack game for a specific guild.
type Game struct {
	guildID       string
	game          *bj.Game
	config        *Config
	state         GameState
	gameStartTime time.Time
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
		guildID: guildID,
		game:    bj.New(config.Decks),
		config:  config,
		state:   NotStarted,
		lock:    sync.Mutex{},
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
	if g.NotStarted() {
		g.state = WaitingForPlayers
	}

	cm := NewChipManager(g.guildID, memberID)
	g.game.AddPlayer(memberID, 0, bj.WithChipManager(cm))
	player := g.GetPlayer(memberID)
	if err := player.PlaceBet(g.config.BetAmount); err != nil {
		g.game.RemovePlayer(memberID)
		return err
	}

	// If this is the first player, set the game start time to wait for additional players.
	if len(g.game.Players()) == 1 {
		g.gameStartTime = time.Now().Add(g.config.WaitForPlayers)
	}

	// If the maximum number of players has been reached, start a new round
	// without waiting for additional players.
	if len(g.game.Players()) == g.config.MaxPlayers {
		g.StartNewRound()
	}
	return nil
}

// GetPlayer retrieves a player from the blackjack game by their member ID.
func (g *Game) GetPlayer(memberID string) *bj.Player {
	return g.game.GetPlayer(memberID)
}

// StartNewRound starts a new round of blackjack in the game.
func (g *Game) StartNewRound() {
	// If the game is already active, do nothing.
	if g.IsActive() {
		return
	}

	g.state = InProgress
	g.game.StartNewRound()
	g.game.DealInitialCards()
}

// EndRound ends the current round of blackjack for the guild, removing all players from the game.
func (g *Game) EndRound() {
	for _, player := range g.game.Players() {
		g.game.RemovePlayer(player.Name())
	}
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

// WaitForPlayers returns the number of seconds remaining to wait for players
// before starting the game. If the wait time has elapsed, it returns 0.
func (g *Game) WaitForPlayers() int {
	waitTime := time.Until(g.gameStartTime)
	if waitTime > 0 {
		return int(waitTime.Seconds())
	}
	g.StartNewRound()
	return 0
}

// Lock locks the game's mutex.
func (g *Game) Lock() {
	g.lock.Lock()
}

// Unlock unlocks the game's mutex.
func (g *Game) Unlock() {
	g.lock.Unlock()
}
