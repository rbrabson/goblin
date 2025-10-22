package blackjack

import (
	"log/slog"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/discord"
)

var (
	waitLock sync.Mutex
	waiting  = make(map[string]*Game)
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"blackjack":   blackjack,
		"join":        blackjackJoinGame,
		"hit":         blackjackHit,
		"stand":       blackjackStand,
		"double_down": blackjackDoubleDown,
		"split":       blackjackSplit,
		"surrender":   blackjackSurrender,
	}

	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "blackjack",
			Description: "Interacts with the blackjack table.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "play",
					Description: "Play the blackjack game.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "stats",
					Description: "Shows a user's stats.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The member or member ID.",
							Required:    false,
						},
					},
				},
			},
		},
	}
)

var (
	joinButton = discordgo.Button{
		Label:    "Join",
		Style:    discordgo.PrimaryButton,
		CustomID: "join",
	}
	hitButton = discordgo.Button{
		Label:    "Hit",
		Style:    discordgo.PrimaryButton,
		CustomID: "hit",
	}
	standButton = discordgo.Button{
		Label:    "Stand",
		Style:    discordgo.PrimaryButton,
		CustomID: "staad",
	}
	doubleDownButton = discordgo.Button{
		Label:    "Double Down",
		Style:    discordgo.PrimaryButton,
		CustomID: "doubleDown",
	}
	splitButton = discordgo.Button{
		Label:    "Split",
		Style:    discordgo.PrimaryButton,
		CustomID: "split",
	}
	surrenderButton = discordgo.Button{
		Label:    "Surrender",
		Style:    discordgo.DangerButton,
		CustomID: "surrender",
	}
)

// blackjack handles the /blackjack command and its subcommands.
func blackjack(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The system is shutting down."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", i.GuildID),
				slog.String("memberID", i.Member.User.ID),
				slog.Any("error", err),
			)
		}
		return
	}

	subCommand := i.ApplicationCommandData().Options[0].Name

	switch subCommand {
	case "play":
		playBlackjack(s, i)
	case "stats":
		showStats(s, i)
	default:
		// Unknown subcommand
	}
}

// playBlackjack handles the /blackjack/play command.
func playBlackjack(s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp := disgomsg.NewResponse(
		disgomsg.WithContent("Not Implemented Yet."),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
	}

	// Add player to the game (handle errors appropriately)

	waitForRoundToStart(s, i)
}

// waitForRoundToStart waits for the round to start for the blackjack game.
func waitForRoundToStart(s *discordgo.Session, i *discordgo.InteractionCreate) {
	waitLock.Lock()

	game, exists := waiting[i.GuildID]
	if exists {
		waitLock.Unlock()
		slog.Info("game already waiting",
			slog.String("guildID", i.GuildID),
		)
		return
	}
	waitLock.Unlock()

	// Wait until either the time expires or the max number of players joins the game
	for {
		if game.IsActive() {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// TODO: Notify players that the round is starting
}

// blackjackJoinGame handles the /blackjack/join command.
func blackjackJoinGame(s *discordgo.Session, i *discordgo.InteractionCreate) {
	game := GetGame(i.GuildID)

	game.Lock()
	defer game.Unlock()

	if !joinChecks(s, i) {
		return
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithContent("Join Not Implemented Yet."),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
	}
}

// blackjackHit handles a player hitting in blackjack.
func blackjackHit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	game := GetGame(i.GuildID)

	game.Lock()
	defer game.Unlock()

	if !playHandChecks(s, i) {
		return
	}

	if err := game.game.PlayerHit(i.Member.User.ID); err != nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Error hitting: " + err.Error()),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", i.GuildID),
				slog.String("memberID", i.Member.User.ID),
				slog.Any("error", err),
			)
		}
		return
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithContent("Hit Not Implemented Yet."),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
	}
}

// blackjackStand handles a player standing in blackjack.
func blackjackStand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	game := GetGame(i.GuildID)

	game.Lock()
	defer game.Unlock()

	if !playHandChecks(s, i) {
		return
	}

	if err := game.game.PlayerStand(i.Member.User.ID); err != nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Error standing: " + err.Error()),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", i.GuildID),
				slog.String("memberID", i.Member.User.ID),
				slog.Any("error", err),
			)
		}
		return
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithContent("Stand Not Implemented Yet."),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
	}
}

// blackjackDoubleDown handles a player doubling down in blackjack.
func blackjackDoubleDown(s *discordgo.Session, i *discordgo.InteractionCreate) {
	game := GetGame(i.GuildID)

	game.Lock()
	defer game.Unlock()

	if !playHandChecks(s, i) {
		return
	}

	if err := game.game.PlayerDoubleDownHit(i.Member.User.ID); err != nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Error doubling down: " + err.Error()),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", i.GuildID),
				slog.String("memberID", i.Member.User.ID),
				slog.Any("error", err),
			)
		}
		return
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithContent("Double Down Not Implemented Yet."),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
	}
}

// blackjackSplit handles a player splitting their hand in blackjack.
func blackjackSplit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	game := GetGame(i.GuildID)

	game.Lock()
	defer game.Unlock()

	if !playHandChecks(s, i) {
		return
	}

	if err := game.game.PlayerSplit(i.Member.User.ID); err != nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Error splitting: " + err.Error()),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", i.GuildID),
				slog.String("memberID", i.Member.User.ID),
				slog.Any("error", err),
			)
		}
		return
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithContent("Split Not Implemented Yet."),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
	}
}

// blackjackSurrender handles a player surrendering in blackjack.
func blackjackSurrender(s *discordgo.Session, i *discordgo.InteractionCreate) {
	game := GetGame(i.GuildID)

	game.Lock()
	defer game.Unlock()

	if !playHandChecks(s, i) {
		return
	}

	if err := game.game.PlayerSurrender(i.Member.User.ID); err != nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Error surrendering: " + err.Error()),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", i.GuildID),
				slog.String("memberID", i.Member.User.ID),
				slog.Any("error", err),
			)
		}
		return
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithContent("Surrender Not Implemented Yet."),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
	}
}

// joinChecks performs checks to see if a player can join the game.
func joinChecks(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	game := GetGame(i.GuildID)
	if game.NotStarted() {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The blackjack game has not started yet. Please wait for the game to start before joining."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send the response",
				slog.Any("error", err),
			)
		}
		slog.Error("blackjack game has not started",
			slog.String("guildID", i.GuildID),
		)
		return false
	}
	if game.IsActive() {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Cannot join an active blackjack game."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send the response",
				slog.Any("error", err),
			)
		}
		slog.Error("blackjack game is already active",
			slog.String("guildID", i.GuildID),
		)
		return false
	}

	player := game.GetPlayer(i.Member.User.ID)
	if player != nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You have already joined the blackjack game."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send the response",
				slog.Any("error", err),
			)
		}
		return false
	}

	return true
}

// playHandChecks performs checks to see if a player can play their hand.
func playHandChecks(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	game := GetGame(i.GuildID)
	if !game.IsActive() {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("There is no active blackjack game. Join the game to start a new round."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send the response",
				slog.Any("error", err),
			)
		}
		slog.Error("blackjack game is not active",
			slog.String("guildID", i.GuildID),
		)
		return false
	}

	player := game.GetPlayer(i.Member.User.ID)
	activePlayer := game.game.GetActivePlayer()
	if player == nil || game.game.GetActivePlayer() != player {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You are not the active player."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send the response",
				slog.Any("error", err),
			)
		}
		slog.Error("not the active blackjack player",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("activePlayer", activePlayer),
		)
		return false
	}

	return true
}

// showStats handles the /blackjack/stats command.
func showStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp := disgomsg.NewResponse(
		disgomsg.WithContent("Not Implemented Yet."),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
	}
}
