package blackjack

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	bj "github.com/rbrabson/blackjack"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/internal/format"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"join_blackjack":       blackjackJoinGame,
		"hit_blackjack":        blackjackHit,
		"stand_blackjack":      blackjackStand,
		"doubledown_blackjack": blackjackDoubleDown,
		"split_blackjack":      blackjackSplit,
		"surrender_blackjack":  blackjackSurrender,
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"blackjack": blackjack,
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
		Style:    discordgo.SuccessButton,
		CustomID: "join_blackjack",
		Emoji:    nil,
	}
	hitButton = discordgo.Button{
		Label:    "Hit",
		Style:    discordgo.PrimaryButton,
		CustomID: "hit_blackjack",
		Emoji:    nil,
	}
	standButton = discordgo.Button{
		Label:    "Stand",
		Style:    discordgo.PrimaryButton,
		CustomID: "stand_blackjack",
		Emoji:    nil,
	}
	doubleDownButton = discordgo.Button{
		Label:    "Double Down",
		Style:    discordgo.PrimaryButton,
		CustomID: "doubledown_blackjack",
		Emoji:    nil,
	}
	splitButton = discordgo.Button{
		Label:    "Split",
		Style:    discordgo.PrimaryButton,
		CustomID: "split_blackjack",
		Emoji:    nil,
	}
	surrenderButton = discordgo.Button{
		Label:    "Surrender",
		Style:    discordgo.DangerButton,
		CustomID: "surrender_blackjack",
		Emoji:    nil,
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

	guild.GetMember(i.GuildID, i.Member.User.ID).SetName(i.Member.User.Username, i.Member.Nick, i.Member.User.GlobalName)

	showJoinGame(s, i, game)
	game.Unlock()

	waitForRoundToStart(s, i, game)

	game.Lock()
	game.StartNewRound()
	showStartingGame(s, game)
	game.Unlock()

	playRound(s, i, game)

	game.EndRound()
}

// waitForRoundToStart waits for the round to start for the blackjack game.
func waitForRoundToStart(s *discordgo.Session, i *discordgo.InteractionCreate, game *Game) {
	// Wait until the game starts or a timeout occurs.
	timeout := time.After(game.config.WaitForPlayers)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			return
		case <-tick:
			secondsBeforeStart := game.SecondsBeforeStart()
			if game.IsActive() || len(game.Players()) == game.config.MaxPlayers || secondsBeforeStart == 0 {
				return
			}
			showJoinGame(s, i, game)
		}
	}
}

// playRound handles playing a round of blackjack.
func playRound(s *discordgo.Session, i *discordgo.InteractionCreate, game *Game) {
	game.DealInitialCards()
	showDeal(s, i, game, false)

	// Check for dealer blackjack, and only proceed to player turns if dealer doesn't have blackjack
	if !game.Dealer().HasBlackjack() {
		playerTurns(s, game)
		dealerTurn(s, i, game)
	}

	for _, player := range game.Players() {
		result := game.EvaluateHand(player.CurrentHand())
		slog.Info("player result",
			slog.String("guildID", game.guildID),
			slog.String("playerName", player.Name()),
			slog.Any("result", result),
		)
	}

	game.PayoutResults()
	showResults(s, game)
}

// playerTurns handles the turns for each player in blackjack, until all players have stood or busted.
func playerTurns(s *discordgo.Session, game *Game) {
	for _, player := range game.Players() {
		playerName := guild.GetMember(game.guildID, player.Name()).Name
		slog.Debug("starting turn for player",
			slog.String("playerName", playerName),
		)
		if !player.IsActive() {
			continue
		}

		for player.HasActiveHands() {
			currentHand := player.CurrentHand()
			currentHandIndex := player.GetCurrentHandNumber()

			slog.Debug("processing player hand",
				slog.String("playerName", playerName),
				slog.Any("hand", currentHand),
			)

			// Check for player blackjack
			if currentHand.IsBlackjack() {
				if !player.MoveToNextActiveHand() {
					player.SetActive(false)
					break
				}
				continue
			}

			// Wait for the player action or timeout
			waitUntil := time.Now().Add(game.config.PlayerTimeout)
			showCurrentTurn(s, game, player, currentHand, currentHandIndex, time.Until(waitUntil))

			// clear turnChan before waiting for action
			for len(game.turnChan) > 0 {
				<-game.turnChan
			}

			var action Action
			timeout := time.After(game.config.PlayerTimeout)
			tick := time.Tick(1 * time.Second)
		GetAction:
			for {
				select {
				case pa := <-game.turnChan:
					action = pa
					slog.Debug("received player action",
						slog.String("guildID", game.guildID),
						slog.String("playerName", playerName),
						slog.Any("action", action),
					)
					break GetAction
				case <-timeout:
					slog.Debug("player turn timed out, defaulting to Stand",
						slog.String("guildID", game.guildID),
						slog.String("playerName", playerName),
					)
					action = Stand
					break GetAction
				case <-tick:
					showCurrentTurn(s, game, player, currentHand, currentHandIndex, time.Until(waitUntil))
				}
			}

			switch action {
			case Hit:
				if err := game.PlayerHit(player.Name()); err != nil {
					slog.Error("error processing player hit",
						slog.String("guildID", game.guildID),
						slog.String("playerName", playerName),
						slog.Any("error", err),
					)
					continue
				}

				if currentHand.IsBusted() {
					slog.Debug("player hand busted",
						slog.String("guildID", game.guildID),
						slog.String("playerName", player.Name()),
					)
					currentHand.SetActive(false)
				}

				if currentHand.Value() == 21 {
					slog.Debug("player hand reached 21",
						slog.String("guildID", game.guildID),
						slog.String("playerName", playerName),
						slog.Any("hand", currentHand),
					)
				} else {
					slog.Debug("hit",
						slog.String("guildID", game.guildID),
						slog.String("playerName", playerName),
						slog.Any("hand", currentHand),
					)
				}

			case Stand:
				if err := game.PlayerStand(player.Name()); err != nil {
					slog.Error("error processing player stand",
						slog.String("guildID", game.guildID),
						slog.String("playerName", playerName),
						slog.Any("error", err),
					)
					continue
				}

				slog.Debug("standing",
					slog.String("guildID", game.guildID),
					slog.String("playerName", playerName),
					slog.Any("hand", currentHand),
				)

			case DoubleDown:
				if !player.CurrentHand().CanDoubleDown() {
					slog.Error("cannot double down",
						slog.String("guildID", game.guildID),
						slog.String("playerName", playerName),
					)
					continue
				}

				if err := player.CurrentHand().DoubleDown(); err != nil {
					slog.Error("error processing player double down",
						slog.String("guildID", game.guildID),
						slog.String("playerName", playerName),
						slog.Any("error", err),
					)
					continue
				}

				if err := game.PlayerDoubleDownHit(player.Name()); err != nil {
					slog.Error("error processing player double down hit",
						slog.String("guildID", game.guildID),
						slog.String("playerName", playerName),
						slog.Any("error", err),
					)
					continue
				}

				if currentHand.IsBusted() {
					slog.Debug("player hand busted after double down",
						slog.String("guildID", game.guildID),
						slog.String("playerName", playerName),
					)
				} else {
					slog.Debug("double down",
						slog.String("guildID", game.guildID),
						slog.String("playerName", playerName),
						slog.Any("hand", currentHand),
					)
				}

			case Split:
				if !player.CurrentHand().CanSplit() {
					slog.Error("cannot split",
						slog.String("guildID", game.guildID),
						slog.String("playerName", playerName),
					)
					continue
				}

				if err := game.PlayerSplit(player.Name()); err != nil {
					slog.Error("error processing player split",
						slog.String("guildID", game.guildID),
						slog.String("playerName", playerName),
						slog.Any("error", err),
					)
					continue
				}

				slog.Debug("split",
					slog.String("guildID", game.guildID),
					slog.String("playerName", playerName),
					slog.Any("hand", currentHand),
				)

			case Surrender:
				if !player.CurrentHand().CanSurrender() {
					slog.Error("cannot surrender",
						slog.String("guildID", game.guildID),
						slog.String("playerName", playerName),
					)
					continue
				}

				if err := game.PlayerSurrender(player.Name()); err != nil {
					slog.Error("error processing player surrender",
						slog.String("guildID", game.guildID),
						slog.String("playerName", playerName),
						slog.Any("error", err),
					)
					continue
				}

				slog.Debug("surrender",
					slog.String("guildID", game.guildID),
					slog.String("playerName", playerName),
					slog.Any("hand", currentHand),
				)

			default:
				slog.Error("invalid player action",
					slog.String("guildID", game.guildID),
					slog.String("playerName", playerName),
					slog.Any("action", action),
				)
			}

			showCurrentTurn(s, game, player, currentHand, currentHandIndex, time.Until(waitUntil))
			time.Sleep(2 * time.Second)
		}

		// Move to next hand if current hand is done
		if !player.CurrentHand().IsActive() {
			if !player.MoveToNextActiveHand() {
				player.SetActive(false)
				continue
			}
		}
	}
}

// dealerTurn handles the dealer's turn in blackjack.
func dealerTurn(s *discordgo.Session, i *discordgo.InteractionCreate, game *Game) {
	slog.Debug("starting dealer turn",
		slog.String("guildID", game.guildID),
	)

	// Dealer turn (if any players are still in)
	if hasNonBustedPlayers(game) {
		showDeal(s, i, game, true)
		time.Sleep(1 * time.Second)
		game.DealerPlay()
		showDeal(s, i, game, true)
		time.Sleep(1 * time.Second)
	}
}

// hasNonBustedPlayers checks if there are any non-busted players in the game.
func hasNonBustedPlayers(game *Game) bool {
	for _, player := range game.Players() {
		for _, hand := range player.Hands() {
			if !(hand.IsBusted() || hand.IsSurrendered() || hand.IsBlackjack()) {
				return true
			}
		}
	}
	return false
}

// showJoinGame displays the join game message with a button to join.
func showJoinGame(s *discordgo.Session, i *discordgo.InteractionCreate, game *Game) {
	p := message.NewPrinter(language.AmericanEnglish)

	seconds := game.SecondsBeforeStart()
	var until string
	if seconds == 1 {
		until = "1 second"
	} else {
		until = p.Sprintf("%d seconds", seconds)
	}

	playerNames := make([]string, 0, len(game.Players()))
	for _, player := range game.Players() {
		member := guild.GetMember(game.guildID, player.Name())
		playerNames = append(playerNames, member.Name)
	}

	embeds := []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Title:       game.symbols["Cards"]["Multiple"] + "Blackjack" + game.symbols["Cards"]["Multiple"],
			Description: p.Sprintf("A new blackjack game is starting. You can join the game for a cost of %d credits at any time prior to the game starting.", game.config.BetAmount),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Status",
					Value:  p.Sprintf("Starts in %s", until),
					Inline: true,
				},
				{
					Name:   p.Sprintf("Players (%d)", len(game.Players())),
					Value:  strings.Join(playerNames, ", "),
					Inline: true,
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
func showStartingGame(s *discordgo.Session, game *Game) {
	p := message.NewPrinter(language.AmericanEnglish)

	playerNames := make([]string, 0, len(game.Players()))
	for _, player := range game.Players() {
		member := guild.GetMember(game.guildID, player.Name())
		playerNames = append(playerNames, member.Name)
	}

	embeds := []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Title:       game.symbols["Cards"]["Multiple"] + "Blackjack" + game.symbols["Cards"]["Multiple"],
			Description: "The round is starting! The dealer is dealing the hands to the players.",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   p.Sprintf("Players (%d)", len(game.Players())),
					Value:  strings.Join(playerNames, ", "),
					Inline: true,
				},
			},
		},
	}

	if _, err := s.InteractionResponseEdit(game.interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &embeds,
		Components: &[]discordgo.MessageComponent{},
	}); err != nil {
		slog.Error("error editing blackjack interaction response",
			slog.String("guildID", game.guildID),
			slog.String("memberID", game.interaction.Member.User.ID),
			slog.Any("error", err),
		)
	}
}

// showDeal displays the deal information for the blackjack game.
func showDeal(s *discordgo.Session, i *discordgo.InteractionCreate, game *Game, isDealerTurn bool) {
	embeds := make([]*discordgo.MessageEmbed, 0, len(game.Players())+1)

	var title string
	if game.message == nil {
		title = game.symbols["Cards"]["Multiple"] + "Blackjack - Deal" + game.symbols["Cards"]["Multiple"]
	} else {
		title = game.symbols["Cards"]["Multiple"] + "Blackjack - Dealer's Turn" + game.symbols["Cards"]["Multiple"]
	}

	dealerHand := game.Dealer().Hand()
	embeds = append(embeds, &discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Title:       title,
		Description: "**Dealer Hand**:\n" + game.symbols.GetHand(dealerHand, !isDealerTurn),
	})

	for _, player := range game.Players() {
		member := guild.GetMember(game.guildID, player.Name())
		playerEmbed := &discordgo.MessageEmbed{
			Type:   discordgo.EmbedTypeRich,
			Title:  member.Name,
			Fields: make([]*discordgo.MessageEmbedField, 0, len(player.Hands())),
		}
		for idx, hand := range player.Hands() {
			handField := &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("Hand %d", idx+1),
				Value:  game.symbols.GetHand(hand, false),
				Inline: false,
			}
			playerEmbed.Fields = append(playerEmbed.Fields, handField)
		}
		embeds = append(embeds, playerEmbed)
	}

	if game.message != nil {
		m, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel:    game.message.ChannelID,
			ID:         game.message.ID,
			Embeds:     &embeds,
			Components: &[]discordgo.MessageComponent{},
		})
		if err != nil {
			slog.Error("error editing blackjack deal message",
				slog.String("guildID", game.guildID),
				slog.String("memberID", game.interaction.Member.User.ID),
				slog.Any("error", err),
			)
			return
		}
		game.message = m
	} else {
		m, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
			Embeds: embeds,
		})
		if err != nil {
			slog.Error("error sending blackjack deal message",
				slog.String("guildID", game.guildID),
				slog.String("memberID", game.interaction.Member.User.ID),
				slog.Any("error", err),
			)
			return
		}
		game.message = m
	}
}

// showCurrentTurn displays the current turn information for the active player.
func showCurrentTurn(s *discordgo.Session, game *Game, currentPlayer *bj.Player, currentHand *bj.Hand, currentHandIndex int, waitTime time.Duration) {
	embeds := make([]*discordgo.MessageEmbed, 0, len(game.Players())+1)

	embeds = append(embeds, &discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Title:       game.symbols["Cards"]["Multiple"] + "Blackjack - Player Turn" + game.symbols["Cards"]["Multiple"],
		Description: "**Dealer Hand**:\n" + game.symbols.GetHand(game.Dealer().Hand(), true),
	})

	for _, player := range game.Players() {
		// If the active player only has a single hand, it will be shown in the active player's turn embed.
		if player == currentPlayer && len(player.Hands()) == 1 {
			continue
		}
		playerEmbed := &discordgo.MessageEmbed{
			Type:   discordgo.EmbedTypeRich,
			Title:  guild.GetMember(game.guildID, player.Name()).Name,
			Fields: make([]*discordgo.MessageEmbedField, 0, len(player.Hands())),
		}
		for idx, hand := range player.Hands() {
			// The active player's current hand is shown in the active player's turn embed.
			if player == currentPlayer && idx == currentHandIndex {
				continue
			}
			handField := &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("Hand %d", idx+1),
				Value:  game.symbols.GetHand(hand, false),
				Inline: false,
			}
			playerEmbed.Fields = append(playerEmbed.Fields, handField)
		}
		embeds = append(embeds, playerEmbed)
	}

	// Buttons for the current player's actions
	buttons := make([]discordgo.MessageComponent, 0, 5)
	// Player actions for current hand.
	if currentHand.IsActive() && !currentHand.IsBusted() && !currentHand.IsBlackjack() {
		buttons = append(buttons, hitButton, standButton)
		if currentHand.CanDoubleDown() {
			buttons = append(buttons, doubleDownButton)
		}
		if currentHand.CanSplit() {
			buttons = append(buttons, splitButton)
		}
		if currentHand.CanSurrender() {
			buttons = append(buttons, surrenderButton)
		}
	}

	// Active player's turn embed
	caser := cases.Title(language.AmericanEnglish)
	actions := strings.ReplaceAll(currentHand.ActionSummary(), ", ", "\n")
	embed := &discordgo.MessageEmbed{
		Type:  discordgo.EmbedTypeRich,
		Title: guild.GetMember(game.guildID, currentPlayer.Name()).Name + "'s Turn",
		Color: 0x00ff00, // Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   fmt.Sprintf("Hand %d", currentHandIndex+1),
				Value:  game.symbols.GetHandWithoutValue(currentHand, false),
				Inline: false,
			},
			{
				Name:   "Actions",
				Value:  actions,
				Inline: false,
			},
			{
				Name:  "Value",
				Value: caser.String(GetHandValue(currentHand, false)),
			},
		},
	}
	if len(buttons) > 0 {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Time remaining: %s", format.Duration(waitTime)),
		}
	}
	embeds = append(embeds, embed)

	var m *discordgo.Message
	var err error

	if len((buttons)) == 0 {
		m, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel:    game.message.ChannelID,
			ID:         game.message.ID,
			Embeds:     &embeds,
			Components: &[]discordgo.MessageComponent{},
		})
	} else {
		m, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel: game.message.ChannelID,
			ID:      game.message.ID,
			Embeds:  &embeds,
			Components: &[]discordgo.MessageComponent{
				discordgo.ActionsRow{Components: buttons},
			},
		})
	}
	if err != nil {
		slog.Error("error editing blackjack turn message",
			slog.String("guildID", game.guildID),
			slog.String("memberID", game.interaction.Member.User.ID),
			slog.Any("error", err),
		)
		return
	}
	game.message = m
}

// showResults displays the results of the blackjack round for each player.
func showResults(s *discordgo.Session, game *Game) {
	p := message.NewPrinter(language.AmericanEnglish)

	embeds := make([]*discordgo.MessageEmbed, 0, len(game.Players())+1)
	embeds = append(embeds, &discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Title:       game.symbols["Cards"]["Multiple"] + "Blackjack - Results" + game.symbols["Cards"]["Multiple"],
		Description: "**Dealer Hand**:\n" + game.symbols.GetHand(game.Dealer().Hand(), false),
	})
	for _, player := range game.Players() {
		embed := &discordgo.MessageEmbed{
			Title:  guild.GetMember(game.guildID, player.Name()).Name,
			Fields: make([]*discordgo.MessageEmbedField, 0, len(player.Hands())),
		}

		for idx, hand := range player.Hands() {
			var result string
			switch {
			case hand.Winnings() > 0:
				if hand.Winnings() == 1 {
					result = "Won 1 credit"
				} else {
					result = p.Sprintf("Won %d credits", hand.Winnings())
				}
			case hand.Winnings() < 0:
				if hand.Winnings() == -1 {
					result = "Lost 1 credit"
				} else {
					result = p.Sprintf("Lost %d credits", -hand.Winnings())

				}
			default:
				result = "Push"
			}
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   p.Sprintf("Hand %d", idx+1),
				Value:  p.Sprintf("%s\n**Result**: %s", game.symbols.GetHand(hand, false), result),
				Inline: false,
			})
		}

		embeds = append(embeds, embed)
	}

	_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    game.message.ChannelID,
		ID:         game.message.ID,
		Embeds:     &embeds,
		Components: &[]discordgo.MessageComponent{},
	})
	if err != nil {
		slog.Error("error sending blackjack result message",
			slog.String("guildID", game.guildID),
			slog.String("memberID", game.interaction.Member.User.ID),
			slog.Any("error", err),
		)
		return
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

	guild.GetMember(i.GuildID, i.Member.User.ID).SetName(i.Member.User.Username, i.Member.Nick, i.Member.User.GlobalName)

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

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
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

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
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

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
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

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
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

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
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
