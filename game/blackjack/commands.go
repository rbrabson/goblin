package blackjack

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/discord"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
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
	game := GetGame(i.GuildID)

	game.Lock()
	if !startChecks(s, i) {
		game.Unlock()
		return
	}
	if err := game.AddPlayer(i.Member.User.ID); err != nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Error starting the game: " + err.Error()),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", i.GuildID),
				slog.String("memberID", i.Member.User.ID),
				slog.Any("error", err),
			)
		}
		game.Unlock()
		return
	}
	showJoinGame(s, i, game)
	game.Unlock()

	waitForRoundToStart(s, i)

	game.Lock()
	game.StartNewRound()
	showStartingGame(s, i, game)
	game.Unlock()

	playRound(s, i)
}

// waitForRoundToStart waits for the round to start for the blackjack game.
func waitForRoundToStart(s *discordgo.Session, i *discordgo.InteractionCreate) {
	game := GetGame(i.GuildID)

	// Wait until the game starts or a timeout occurs.
	timeout := time.After(game.config.WaitForPlayers)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			return
		case <-tick:
			secondsBeforeStart := game.SecondsBeforeStart()
			if game.IsActive() || secondsBeforeStart == 0 {
				return
			}
			showJoinGame(s, i, game)
		}
	}
}

// playRound handles playing a round of blackjack.
func playRound(s *discordgo.Session, i *discordgo.InteractionCreate) {
	game := GetGame(i.GuildID)
	defer game.EndRound()

	game.StartNewRound()
	game.DealInitialCards()
	showDeal(s, i, game)

	// Check for dealer blackjack, and only proceed to player turns if dealer doesn't have blackjack
	if !game.Dealer().HasBlackjack() {
		playerTurns(s, i)
		dealerTurn(s, i)
	}

	showResults(game)
	game.PayoutResults()
}

// playerTurns handles the turns for each player in blackjack, until all players have stood or busted.
func playerTurns(s *discordgo.Session, i *discordgo.InteractionCreate) {
	game := GetGame(i.GuildID)
	for _, player := range game.Players() {
		if !player.IsActive() {
			continue
		}

		for player.HasActiveHands() {
			currentHand := player.CurrentHand()

			// Check for player blackjack
			if currentHand.IsBlackjack() {
				fmt.Printf("🎯 %s has blackjack on hand %d!\n", player.Name(), player.GetCurrentHandIndex()+1)
				if !player.MoveToNextActiveHand() {
					player.SetActive(false)
					break
				}
				continue
			}

			showCurrentTurn(s, i, game)

			// TODO: use a Select with a timeout here; if time out, then select "Stand" for the player action
			action := <-game.turnChan

			switch action {
			case Hit:
				err := game.PlayerHit(player.Name())
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					continue
				}

				fmt.Printf("Drew: %s\n", currentHand.String())

				if currentHand.IsBusted() {
					fmt.Printf("💥 Hand busted!\n")
					currentHand.SetActive(false)
				}

			case Stand:
				fmt.Printf("Standing on hand.\n")
				err := game.PlayerStand(player.Name())
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					continue
				}

			case DoubleDown:
				if !player.CanDoubleDown() {
					fmt.Println("Cannot double down.")
					continue
				}

				err := player.DoubleDown()
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					continue
				}

				err = game.PlayerDoubleDownHit(player.Name())
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					continue
				}

				fmt.Printf("Doubled down! Drew: %s\n", currentHand.String())

				if currentHand.IsBusted() {
					fmt.Printf("💥 Hand busted!\n")
				}

				// Double down ends the hand
				err = game.PlayerStand(player.Name())
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}

			case Split:
				if !player.CanSplit() {
					fmt.Println("Cannot split.")
					continue
				}

				err := game.PlayerSplit(player.Name())
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					continue
				}

				fmt.Printf("Hand split! You now have %d hands.\n", len(player.Hands()))
				// Show current hand after split
				fmt.Printf("Current hand: %s\n", currentHand.String())

			case Surrender:
				if !player.CanSurrender() {
					fmt.Println("Cannot surrender.")
					continue
				}

				err := game.PlayerSurrender(player.Name())
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					continue
				}

				fmt.Printf("Surrendered! Half bet returned.\n")

			default:
				fmt.Println("Invalid action. Please choose (h)it, (s)tand, (d)ouble down, s(p)lit, or s(u)rrender if available.")
			}
		}

		// Move to next hand if current hand is done
		if !player.CurrentHand().IsActive() {
			if !player.MoveToNextActiveHand() {
				player.SetActive(false)
				break
			}
		}
	}
}

// dealerTurn handles the dealer's turn in blackjack.
func dealerTurn(s *discordgo.Session, i *discordgo.InteractionCreate) {
	game := GetGame(i.GuildID)

	// Dealer turn (if any players are still in)
	if hasActiveNonBustedPlayers(game) {
		fmt.Println("\n🎯 Dealer's turn:")
		fmt.Println("Revealing hole card...")
		fmt.Println(game.Dealer().RevealHoleCard())

		game.DealerPlay()
	}

	for _, action := range game.Dealer().Hand().Actions() {
		fmt.Println(action)
	}
}

// hasActiveNonBustedPlayers checks if there are any active non-busted players in the game.
func hasActiveNonBustedPlayers(game *Game) bool {
	for _, player := range game.Players() {
		if player.Bet() > 0 && !player.CurrentHand().IsBusted() {
			return true
		}
	}
	return false
}

// showJoinGame displays the join game message with a button to join.
func showJoinGame(s *discordgo.Session, i *discordgo.InteractionCreate, game *Game) {
	msg := "A new game of blackjack has started! Click the button below to join the game."
	p := message.NewPrinter(language.AmericanEnglish)

	playerNames := make([]string, 0, len(game.Players()))
	for _, player := range game.Players() {
		playerNames = append(playerNames, "<@"+player.Name()+">")
	}

	embeds := []*discordgo.MessageEmbed{
		{
			Type:  discordgo.EmbedTypeRich,
			Title: "Blackjack",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   msg,
					Inline: false,
				},
				{
					Name:   p.Sprintf("Players (%d)", len(game.Players())),
					Value:  strings.Join(playerNames, ", "),
					Inline: false,
				},
			},
		},
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				joinButton,
			},
		},
	}
	if game.interaction == nil {
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds:     embeds,
				Components: components,
			},
		}); err != nil {
			slog.Error("error sending blackjack interaction response",
				slog.String("guildID", game.guildID),
				slog.String("memberID", game.interaction.Member.User.ID),
				slog.Any("error", err),
			)
		}
		game.interaction = i
	} else {
		if _, err := s.InteractionResponseEdit(game.interaction.Interaction, &discordgo.WebhookEdit{
			Embeds:     &embeds,
			Components: &components,
		}); err != nil {
			slog.Error("error editing blackjack interaction response",
				slog.String("guildID", game.guildID),
				slog.String("memberID", game.interaction.Member.User.ID),
				slog.Any("error", err),
			)
		}
	}
}

// showStartingGame displays the starting game message when the round begins.
func showStartingGame(s *discordgo.Session, i *discordgo.InteractionCreate, game *Game) {
	msg := "A new game of blackjack is starting! The dealer is dealing the hands!"
	p := message.NewPrinter(language.AmericanEnglish)

	playerNames := make([]string, 0, len(game.Players()))
	for _, player := range game.Players() {
		playerNames = append(playerNames, "<@"+player.Name()+">")
	}

	embeds := []*discordgo.MessageEmbed{
		{
			Type:  discordgo.EmbedTypeRich,
			Title: "Blackjack",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   msg,
					Inline: false,
				},
				{
					Name:   p.Sprintf("Players (%d)", len(game.Players())),
					Value:  strings.Join(playerNames, ", "),
					Inline: false,
				},
			},
		},
	}

	if _, err := s.InteractionResponseEdit(game.interaction.Interaction, &discordgo.WebhookEdit{
		Embeds: &embeds,
	}); err != nil {
		slog.Error("error editing blackjack interaction response",
			slog.String("guildID", game.guildID),
			slog.String("memberID", game.interaction.Member.User.ID),
			slog.Any("error", err),
		)
	}
}

// showDeal displays the deal information for the blackjack game.
func showDeal(s *discordgo.Session, i *discordgo.InteractionCreate, game *Game) {
	var status strings.Builder

	// Show the dealer's hand (with one card hidden)
	status.WriteString(fmt.Sprintf("%s\n", game.Dealer().StringHidden()))

	// Show players
	for _, player := range game.Players() {
		status.WriteString(fmt.Sprintf("%s\n", player.String()))
	}

	fmt.Println(status.String())
}

// showCurrentTurn displays the current turn information for the active player.
func showCurrentTurn(s *discordgo.Session, i *discordgo.InteractionCreate, game *Game) {
	for _, player := range game.Players() {
		if player != game.GetActivePlayer() {
			currentHand := player.CurrentHand()
			for _, hand := range player.Hands() {
				if len(player.Hands()) > 1 {
					fmt.Printf("\n%s - Hand %d of %d: %s",
						player.Name(),
						player.GetCurrentHandIndex()+1,
						len(player.Hands()),
						hand.String())
				} else {
					fmt.Printf("\n%s: %s", player.Name(), hand.String())
				}
				if &hand == currentHand {
					fmt.Print("  <-- Current Hand\n")
				} else {
					fmt.Print("\n")
				}
			}
		}
	}

	// Player actions for current hand. These should be adding buttons to the message.
	player := game.GetActivePlayer()
	currentHand := player.CurrentHand()
	if currentHand.IsActive() && !currentHand.IsBusted() && !currentHand.IsBlackjack() {
		fmt.Printf("\n%s's turn - Current hand: %s\n", player.Name(), currentHand.String())
		fmt.Print("Choose action: ")
		fmt.Print("(h)it")
		fmt.Print(", (s)tand")
		if player.CanDoubleDown() {
			fmt.Print(", (d)ouble down")
		}
		if player.CanSplit() {
			fmt.Print(", s(p)lit")
		}
		if player.CanSurrender() {
			fmt.Print(", s(u)rrender")
		}
	}
}

// showResults displays the results of the blackjack round for each player.
func showResults(game *Game) {
	fmt.Println("\n💰 Round Results:")
	fmt.Println("================")

	for _, player := range game.Players() {
		if player.Bet() == 0 {
			continue
		}

		hands := player.Hands()
		if len(hands) == 1 {
			// Single hand
			result := game.EvaluateHand(player)
			fmt.Printf("%s: %s\n", player.Name(), result.String())
		} else {
			// Multiple hands (splits)
			fmt.Printf("%s:\n", player.Name())
			for i := 0; i < len(hands); i++ {
				// Temporarily set current hand for evaluation
				originalHandIdx := player.GetCurrentHandIndex()
				player.SetCurrentHandIndex(i)
				result := game.EvaluateHand(player)
				fmt.Printf("  Hand %d: %s\n", i+1, result.String())
				player.SetCurrentHandIndex(originalHandIdx)
			}
		}
		fmt.Printf("  Final Chips: %d\n", player.Chips())
	}
}

// blackjackJoinGame handles the /blackjack/join command.
func blackjackJoinGame(s *discordgo.Session, i *discordgo.InteractionCreate) {
	game := GetGame(i.GuildID)

	game.Lock()
	defer game.Unlock()

	if !joinChecks(s, i) {
		return
	}

	if err := game.AddPlayer(i.Member.User.ID); err != nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Error joining the game: " + err.Error()),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", i.GuildID),
				slog.String("memberID", i.Member.User.ID),
				slog.Any("error", err),
			)
		}
		slog.Error("error adding player to blackjack game",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
		return
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithContent("You have joined the game."),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
	}

	showJoinGame(s, i, game)
}

// blackjackHit handles a player hitting in blackjack.
func blackjackHit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	game := GetGame(i.GuildID)

	game.Lock()
	defer game.Unlock()

	if !playHandChecks(s, i) {
		return
	}

	game.turnChan <- Hit
}

// blackjackStand handles a player standing in blackjack.
func blackjackStand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	game := GetGame(i.GuildID)

	game.Lock()
	defer game.Unlock()

	if !playHandChecks(s, i) {
		return
	}

	game.turnChan <- Stand
}

// blackjackDoubleDown handles a player doubling down in blackjack.
func blackjackDoubleDown(s *discordgo.Session, i *discordgo.InteractionCreate) {
	game := GetGame(i.GuildID)

	game.Lock()
	defer game.Unlock()

	if !playHandChecks(s, i) {
		return
	}

	game.turnChan <- DoubleDown
}

// blackjackSplit handles a player splitting their hand in blackjack.
func blackjackSplit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	game := GetGame(i.GuildID)

	game.Lock()
	defer game.Unlock()

	if !playHandChecks(s, i) {
		return
	}

	game.turnChan <- Split
}

// blackjackSurrender handles a player surrendering in blackjack.
func blackjackSurrender(s *discordgo.Session, i *discordgo.InteractionCreate) {
	game := GetGame(i.GuildID)

	game.Lock()
	defer game.Unlock()

	if !playHandChecks(s, i) {
		return
	}

	game.turnChan <- Surrender
}

// startChecks performs checks to see if a game can be started.
func startChecks(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	game := GetGame(i.GuildID)

	if !game.NotStarted() {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("An active blackjack game already exists."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send the response",
				slog.Any("error", err),
			)
		}
		slog.Error("blackjack game has not started",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
		)
		return false
	}
	return true
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
	activePlayer := game.GetActivePlayer()
	if player == nil || player != activePlayer {
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
