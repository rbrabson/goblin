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
	"github.com/rbrabson/goblin/internal/unicode"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"join_blackjack":       blackjackJoin,
		"hit_blackjack":        blackjackHit,
		"stand_blackjack":      blackjackStand,
		"doubledown_blackjack": blackjackDoubleDown,
		"split_blackjack":      blackjackSplit,
		"surrender_blackjack":  blackjackSurrender,
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"blackjack":       blackjack,
		"blackjack-admin": blackjackAdmin,
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

	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "blackjack-admin",
			Description: "Configures the blackjack game.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "config",
					Description: "Configures the blackjack game.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "info",
							Description: "Returns the configuration information for the server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
						{
							Name:        "bet",
							Description: "Sets the bet amount.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "amount",
									Description: "The amount to set the bet to.",
									Required:    true,
								},
							},
						},
						{
							Name:        "payout",
							Description: "The base payout percentage when winning a game.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "percent",
									Description: "The amount to set the payout percentage to.",
									Required:    true,
								},
							},
						},
						{
							Name:        "single-player",
							Description: "Controls whether the game is single-player only or allows multiple players to join.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionBoolean,
									Name:        "enabled",
									Description: "Whether single-player mode is enabled.",
									Required:    true,
								},
							},
						},
					},
				},
			},
		},
	}
)

// blackjackAdmin handles the /blackjack-admin command and its subcommands.
func blackjackAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		disgomsg.NewResponse(disgomsg.WithContent("The system is shutting down.")).SendEphemeral(s, i.Interaction)
		return
	}

	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		disgomsg.NewResponse(disgomsg.WithContent("You do not have permission to use this command.")).SendEphemeral(s, i.Interaction)
		return
	}

	options := i.ApplicationCommandData().Options
	if options[0].Name == "config" {
		config(s, i)
	}
}

// config handles the /blackjack-admin config command and its subcommands.
func config(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options[0].Options
	switch options[0].Name {
	case "bet":
		configBetAmount(s, i)
	case "payout":
		configPayoutPercent(s, i)
	case "single-player":
		configSinglePlayer(s, i)
	case "info":
		configInfo(s, i)
	}
}

// configBetAmount sets the bet amount for the blackjack game on this server.
func configBetAmount(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := GetConfig(i.GuildID)
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	betAmount := options[0].IntValue()
	config.BetAmount = int(betAmount)

	p := message.NewPrinter(language.AmericanEnglish)
	disgomsg.NewResponse(disgomsg.WithContent(p.Sprintf("Bet amount set to %d", betAmount))).Send(s, i.Interaction)
	writeConfig(config)
	slog.Info("blackjack bet amount updated", slog.String("guildID", i.GuildID), slog.Int("betAmount", int(betAmount)))
}

// configPayoutPercent sets the payout percent for the blackjack game on this server.
func configPayoutPercent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := GetConfig(i.GuildID)
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	payoutPercent := options[0].IntValue()
	config.PayoutPercent = int(payoutPercent)

	p := message.NewPrinter(language.AmericanEnglish)
	disgomsg.NewResponse(disgomsg.WithContent(p.Sprintf("Payout percent set to %d", payoutPercent))).Send(s, i.Interaction)
	writeConfig(config)
	slog.Info("blackjack payout percent updated", slog.String("guildID", i.GuildID), slog.Int("payoutPercent", int(payoutPercent)))
}

// configSinglePlayer sets the single-player mode for the blackjack game on this server.
func configSinglePlayer(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := GetConfig(i.GuildID)
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	singlePlayer := options[0].BoolValue()
	config.SinglePlayerMode = singlePlayer

	p := message.NewPrinter(language.AmericanEnglish)
	disgomsg.NewResponse(disgomsg.WithContent(p.Sprintf("Single-player mode set to %t", singlePlayer))).Send(s, i.Interaction)
	writeConfig(config)
	slog.Info("blackjack single-player mode updated", slog.String("guildID", i.GuildID), slog.Bool("singlePlayerMode", singlePlayer))
}

// configInfo returns the configuration for the blackjack game on this server.
func configInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := GetConfig(i.GuildID)

	embed := &discordgo.MessageEmbed{
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "bet amount",
				Value:  fmt.Sprintf("%d", config.BetAmount),
				Inline: true,
			},
			{
				Name:   "payout percent",
				Value:  fmt.Sprintf("%d", config.PayoutPercent),
				Inline: true,
			},
			{
				Name:   "single player",
				Value:  fmt.Sprintf("%t", config.SinglePlayerMode),
				Inline: true,
			},
		},
	}

	embeds := []*discordgo.MessageEmbed{
		embed,
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Blackjack Configuration",
			Embeds:  embeds,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// blackjack handles the /blackjack command and its subcommands.
func blackjack(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		disgomsg.NewResponse(disgomsg.WithContent("The system is shutting down.")).SendEphemeral(s, i.Interaction)
		return
	}

	subCommand := i.ApplicationCommandData().Options[0].Name

	switch subCommand {
	case "play":
		playBlackjack(s, i)
	case "stats":
		showStats(s, i)
	}
}

// playBlackjack handles the /blackjack/play command.
func playBlackjack(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guild.GetMember(i.GuildID, i.Member.User.ID).SetName(i.Member.User.Username, i.Member.Nick, i.Member.User.GlobalName)

	uid := getUID(i.GuildID, i.Member.User.ID)
	game := GetGame(i.GuildID, uid)

	if err := game.Start(i.Member.User.ID); err != nil {
		disgomsg.NewResponse(disgomsg.WithContent(unicode.FirstToUpper(err.Error()))).SendEphemeral(s, i.Interaction)
		return
	}
	defer game.EndRound()

	showJoinGame(s, i, game)
	waitForRoundToStart(s, i, game)

	if err := game.StartNewRound(); err != nil {
		disgomsg.NewResponse(disgomsg.WithContent(unicode.FirstToUpper(err.Error()))).Send(s, i.Interaction)
		return
	}
	showStartingGame(s, i, game)

	playRound(s, i, game)
}

// waitForRoundToStart waits for the round to start for the blackjack game.
func waitForRoundToStart(s *discordgo.Session, i *discordgo.InteractionCreate, game *Game) {
	if game.config.SinglePlayerMode {
		return
	}

	memberCanJoin := time.Now().Add(game.config.WaitForPlayers)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	count := 0
	showJoinGame(s, i, game)
	for range ticker.C {
		count++

		if game.IsActive() {
			break
		}
		if time.Until(memberCanJoin) <= 0 {
			break
		}
		if len(game.Players()) >= game.config.MaxPlayers {
			break
		}

		if count%5 == 0 {
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
		allPlayerTurns(s, game)
		dealerTurn(s, i, game)
	}

	for _, player := range game.Players() {
		result := game.EvaluateHand(player.CurrentHand())
		slog.Debug("player result",
			slog.String("guildID", game.guildID),
			slog.String("playerName", player.Name()),
			slog.Any("result", result),
		)
	}

	game.PayoutResults()
	showResults(s, game)
}

// allPlayerTurns handles the turns for each player in blackjack, until all players have stood or busted.
func allPlayerTurns(s *discordgo.Session, game *Game) {
	for _, player := range game.Players() {
		playerTurn(s, game, player)
	}
}

// playerTurn handles the turns for a given player in blackjack, until they have stood or busted on their all hands.
func playerTurn(s *discordgo.Session, game *Game, player *bj.Player) {
	playerName := guild.GetMember(game.guildID, player.Name()).Name
	slog.Debug("starting turn for player", slog.String("playerName", playerName))

	if !player.IsActive() {
		return
	}

	for player.HasActiveHands() {
		currentHand := player.CurrentHand()
		playHand(s, game, player, currentHand)

		// Move to next hand if current hand is done
		if !player.CurrentHand().IsActive() {
			if !player.MoveToNextActiveHand() {
				player.SetActive(false)
			}
		}
	}
}

// playHand handles the turn for a specific hand of a player in blackjack, until they have stood or busted on that hand.
func playHand(s *discordgo.Session, game *Game, player *bj.Player, currentHand *bj.Hand) {
	currentHandIndex := player.GetCurrentHandNumber()
	playerName := guild.GetMember(game.guildID, player.Name()).Name

	slog.Debug("processing player hand", slog.String("playerName", playerName), slog.Any("hand", currentHand))

	if currentHand.IsBlackjack() {
		return
	}

	// Wait for the player action or timeout
	waitUntil := time.Now().Add(game.config.PlayerTimeout)
	showCurrentTurn(s, game, player, currentHand, currentHandIndex, time.Until(waitUntil))

	var action Action
	timeout := time.After(game.config.PlayerTimeout)
	tick := time.Tick(1 * time.Second)
GetAction:
	for {
		select {
		case pa := <-game.turnChan:
			action = pa
			slog.Debug("received player action", slog.String("guildID", game.guildID), slog.String("playerName", playerName), slog.Any("action", action))
			break GetAction
		case <-timeout:
			slog.Debug("player turn timed out, defaulting to Stand", slog.String("guildID", game.guildID), slog.String("playerName", playerName))
			action = Stand
			break GetAction
		case <-tick:
			showCurrentTurn(s, game, player, currentHand, currentHandIndex, time.Until(waitUntil))
		}
	}

	// Process the player's move
	switch action {
	case Hit:
		if err := game.PlayerHit(player); err != nil {
			slog.Error("error processing player hit", slog.String("guildID", game.guildID), slog.String("playerName", playerName), slog.Any("error", err))
			return
		}

		if currentHand.IsBusted() {
			slog.Debug("player hand busted", slog.String("guildID", game.guildID), slog.String("playerName", playerName))
			currentHand.SetActive(false)
		}

		if currentHand.Value() == 21 {
			slog.Debug("player hand reached 21", slog.String("guildID", game.guildID), slog.String("playerName", playerName), slog.Any("hand", currentHand))
		} else {
			slog.Debug("hit", slog.String("guildID", game.guildID), slog.String("playerName", playerName), slog.Any("hand", currentHand))
		}

	case Stand:
		if err := game.PlayerStand(player); err != nil {
			slog.Error("error processing player stand", slog.String("guildID", game.guildID), slog.String("playerName", playerName), slog.Any("error", err))
			return
		}

		slog.Debug("standing", slog.String("guildID", game.guildID), slog.String("playerName", playerName), slog.Any("hand", currentHand))

	case DoubleDown:
		if err := game.PlayerDoubleDown(player); err != nil {
			slog.Error("error processing player double down", slog.String("guildID", game.guildID), slog.String("playerName", playerName), slog.Any("error", err))
			return
		}

		if currentHand.IsBusted() {
			slog.Debug("player hand busted after double down", slog.String("guildID", game.guildID), slog.String("playerName", playerName))
		} else {
			slog.Debug("double down", slog.String("guildID", game.guildID), slog.String("playerName", playerName), slog.Any("hand", currentHand))
		}

	case Split:
		if err := game.PlayerSplit(player); err != nil {
			slog.Error("error processing player split", slog.String("guildID", game.guildID), slog.String("playerName", playerName), slog.Any("error", err))
			return
		}

		slog.Debug("split", slog.String("guildID", game.guildID), slog.String("playerName", playerName), slog.Any("hand", currentHand))

	case Surrender:
		if err := game.PlayerSurrender(player); err != nil {
			slog.Error("error processing player surrender", slog.String("guildID", game.guildID), slog.String("playerName", playerName), slog.Any("error", err))
			return
		}

		slog.Debug("surrender", slog.String("guildID", game.guildID), slog.String("playerName", playerName), slog.Any("hand", currentHand))

	default:
		slog.Error("invalid player action", slog.String("guildID", game.guildID), slog.String("playerName", playerName), slog.Any("action", action))
	}

	showCurrentTurn(s, game, player, currentHand, currentHandIndex, time.Until(waitUntil))
	time.Sleep(game.config.ShowPlayerTurn)
}

// dealerTurn handles the dealer's turn in blackjack.
func dealerTurn(s *discordgo.Session, i *discordgo.InteractionCreate, game *Game) {
	slog.Debug("starting dealer turn", slog.String("guildID", game.guildID))

	if err := game.DealerPlay(); err != nil {
		showDeal(s, i, game, true)
		time.Sleep(game.config.ShowDealerTurn)
	}
}

// showJoinGame displays the join game message with a button to join.
func showJoinGame(s *discordgo.Session, i *discordgo.InteractionCreate, game *Game) {
	if game.config.SinglePlayerMode {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Starting single-player blackjack game...",
			},
		})
		return
	}

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
				game.joinButton,
			},
		},
	}
	if game.interaction == nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds:     embeds,
				Components: components,
			},
		})
		game.interaction = i
	} else {
		s.InteractionResponseEdit(game.interaction.Interaction, &discordgo.WebhookEdit{
			Embeds:     &embeds,
			Components: &components,
		})
	}
}

// showStartingGame displays the starting game message when the round begins.
func showStartingGame(s *discordgo.Session, i *discordgo.InteractionCreate, game *Game) {
	if game.config.SinglePlayerMode {
		return
	}

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

	if game.interaction == nil {
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: embeds,
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
			Components: &[]discordgo.MessageComponent{},
		}); err != nil {
			slog.Error("error editing blackjack interaction response",
				slog.String("guildID", game.guildID),
				slog.String("memberID", game.interaction.Member.User.ID),
				slog.Any("error", err),
			)
		}
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
		Description: fmt.Sprintf("**Dealer Hand**:\n%s\nValue: %s", game.symbols.GetHandWithoutValue(dealerHand, !isDealerTurn), GetHandValue(dealerHand, !isDealerTurn)),
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
				Value:  fmt.Sprintf("%s\nValue: %s", game.symbols.GetHandWithoutValue(hand, false), GetHandValue(hand, false)),
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

	dealerHand := game.Dealer().Hand()
	embeds = append(embeds, &discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Title:       game.symbols["Cards"]["Multiple"] + "Blackjack - Player Turn" + game.symbols["Cards"]["Multiple"],
		Description: fmt.Sprintf("**Dealer Hand**:\n%s\nValue: %s", game.symbols.GetHandWithoutValue(dealerHand, true), GetHandValue(dealerHand, true)),
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
				Value:  fmt.Sprintf("%s\nValue: %s", game.symbols.GetHandWithoutValue(hand, false), GetHandValue(hand, false)),
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
		buttons = append(buttons, game.hitButton, game.standButton)
		if currentHand.CanDoubleDown() {
			buttons = append(buttons, game.doubleDownButton)
		}
		if currentHand.CanSplit() {
			buttons = append(buttons, game.splitButton)
		}
		if currentHand.CanSurrender() {
			buttons = append(buttons, game.surrenderButton)
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
			Text: fmt.Sprintf("%s remaining to take an action", format.Duration(waitTime)),
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
	dealerHand := game.Dealer().Hand()
	embeds = append(embeds, &discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Title:       game.symbols["Cards"]["Multiple"] + "Blackjack - Results" + game.symbols["Cards"]["Multiple"],
		Description: fmt.Sprintf("**Dealer Hand**:\n%s\nValue: %s", game.symbols.GetHandWithoutValue(dealerHand, false), GetHandValue(dealerHand, false)),
	})
	for _, player := range game.Players() {
		embed := &discordgo.MessageEmbed{
			Title:  guild.GetMember(game.guildID, player.Name()).Name,
			Fields: make([]*discordgo.MessageEmbedField, 0, len(player.Hands())),
		}

		for idx, hand := range player.Hands() {
			var result string
			winnings := hand.Winnings()
			switch {
			case winnings > 0:
				winnings = winnings * game.config.PayoutPercent / 100
				if winnings == 1 {
					result = "Won 1 credit"
				} else {
					result = p.Sprintf("Won %d credits", winnings)
				}
			case winnings < 0:
				if winnings == -1 {
					result = "Lost 1 credit"
				} else {
					result = p.Sprintf("Lost %d credits", -winnings)

				}
			default:
				result = "Push"
			}
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:  p.Sprintf("Hand %d", idx+1),
				Value: fmt.Sprintf("%s\nValue: %s\n%s", game.symbols.GetHandWithoutValue(hand, false), GetHandValue(hand, false), result),

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

// blackjackJoin handles the /blackjack/join command.
func blackjackJoin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guild.GetMember(i.GuildID, i.Member.User.ID).SetName(i.Member.User.Username, i.Member.Nick, i.Member.User.GlobalName)

	uid := getUIDFromInteraction(i)
	game := GetGame(i.GuildID, uid)

	if err := game.joinGame(i.Member.User.ID); err != nil {
		disgomsg.NewResponse(disgomsg.WithContent(unicode.FirstToUpper(err.Error()))).SendEphemeral(s, i.Interaction)
		slog.Error("error adding player to blackjack game", slog.String("guildID", i.GuildID), slog.String("memberID", i.Member.User.ID), slog.Any("error", err))
		return
	}
	disgomsg.NewResponse(disgomsg.WithContent("You have joined the game.")).SendEphemeral(s, i.Interaction)
	showJoinGame(s, i, game)
}

// blackjackHit handles a player hitting in blackjack.
func blackjackHit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	uid := getUIDFromInteraction(i)
	game := GetGame(i.GuildID, uid)

	if err := game.PlayerActionRequest(i.Member.User.ID, Hit); err != nil {
		disgomsg.NewResponse(disgomsg.WithContent(unicode.FirstToUpper(err.Error()))).SendEphemeral(s, i.Interaction)
		slog.Error("error processing player hit request", slog.String("guildID", game.guildID), slog.String("memberID", i.Member.User.ID), slog.Any("error", err))
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
}

// blackjackStand handles a player standing in blackjack.
func blackjackStand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	uid := getUIDFromInteraction(i)
	game := GetGame(i.GuildID, uid)

	if err := game.PlayerActionRequest(i.Member.User.ID, Stand); err != nil {
		disgomsg.NewResponse(disgomsg.WithContent(unicode.FirstToUpper(err.Error()))).SendEphemeral(s, i.Interaction)
		slog.Error("error processing player stand request", slog.String("guildID", game.guildID), slog.String("memberID", i.Member.User.ID), slog.Any("error", err))
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
}

// blackjackDoubleDown handles a player doubling down in blackjack.
func blackjackDoubleDown(s *discordgo.Session, i *discordgo.InteractionCreate) {
	uid := getUIDFromInteraction(i)
	game := GetGame(i.GuildID, uid)

	if err := game.PlayerActionRequest(i.Member.User.ID, DoubleDown); err != nil {
		disgomsg.NewResponse(disgomsg.WithContent(unicode.FirstToUpper(err.Error()))).SendEphemeral(s, i.Interaction)
		slog.Error("error processing player double down request", slog.String("guildID", game.guildID), slog.String("memberID", i.Member.User.ID), slog.Any("error", err))
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
}

// blackjackSplit handles a player splitting their hand in blackjack.
func blackjackSplit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	uid := getUIDFromInteraction(i)
	game := GetGame(i.GuildID, uid)

	if err := game.PlayerActionRequest(i.Member.User.ID, Split); err != nil {
		disgomsg.NewResponse(disgomsg.WithContent(unicode.FirstToUpper(err.Error()))).SendEphemeral(s, i.Interaction)
		slog.Error("error processing player split request", slog.String("guildID", game.guildID), slog.String("memberID", i.Member.User.ID), slog.Any("error", err))
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
}

// blackjackSurrender handles a player surrendering in blackjack.
func blackjackSurrender(s *discordgo.Session, i *discordgo.InteractionCreate) {
	uid := getUIDFromInteraction(i)
	game := GetGame(i.GuildID, uid)

	if err := game.PlayerActionRequest(i.Member.User.ID, Surrender); err != nil {
		disgomsg.NewResponse(disgomsg.WithContent(unicode.FirstToUpper(err.Error()))).SendEphemeral(s, i.Interaction)
		slog.Error("error processing player surrender request", slog.String("guildID", game.guildID), slog.String("memberID", i.Member.User.ID), slog.Any("error", err))
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
}

// showStats handles the /blackjack/stats command.
func showStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	// Determine which user's stats to show
	var targetUserID string
	var targetUser *discordgo.User

	options := i.ApplicationCommandData().Options[0].Options
	if len(options) > 0 && options[0].Name == "user" {
		targetUser = options[0].UserValue(s)
		targetUserID = targetUser.ID
	} else {
		targetUserID = i.Member.User.ID
	}

	// Get member statistics
	member := GetMember(i.GuildID, targetUserID)

	// Calculate derived statistics
	totalGames := member.Wins + member.Losses + member.Pushes
	winRate := 0.0
	netCredits := member.CreditsWon - member.CreditsLost

	if totalGames > 0 {
		winRate = (float64(member.Wins) / float64(totalGames)) * 100
	}

	// Determine user display name
	var displayName string
	if i.Member.User.ID == targetUserID {
		displayName = "Your"
	} else {
		// Try to get the member from the guild to get their display name
		guildMember, err := s.GuildMember(i.GuildID, targetUserID)
		if err == nil {
			switch {
			case guildMember.Nick != "":
				displayName = guildMember.Nick + "'s"
			case guildMember.User.GlobalName != "":
				displayName = guildMember.User.GlobalName + "'s"
			default:
				displayName = targetUser.Username + "'s"
			}
		} else {
			displayName = targetUser.Username + "'s"
		}
	}

	// Create the stats embed
	embed := &discordgo.MessageEmbed{
		Type:  discordgo.EmbedTypeRich,
		Title: fmt.Sprintf("🃏 %s Blackjack Statistics", displayName),
		Color: 0x2f3136, // Dark gray color
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "📊 Game Summary",
				Value: fmt.Sprintf("**Rounds Played:** %s\n**Hands Played:** %s\n**Win Rate:** %.1f%%",
					p.Sprintf("%d", member.RoundsPlayed),
					p.Sprintf("%d", member.HandsPlayed),
					winRate),
				Inline: false,
			},
			{
				Name: "🎯 Hand Results",
				Value: fmt.Sprintf("**Wins:** %s\n**Losses:** %s\n**Pushes:** %s",
					p.Sprintf("%d", member.Wins),
					p.Sprintf("%d", member.Losses),
					p.Sprintf("%d", member.Pushes)),
				Inline: true,
			},
			{
				Name: "🎴 Special Hands",
				Value: fmt.Sprintf("**Blackjacks:** %s\n**Splits:** %s\n**Surrenders:** %s",
					p.Sprintf("%d", member.Blackjacks),
					p.Sprintf("%d", member.Splits),
					p.Sprintf("%d", member.Surrenders)),
				Inline: true,
			},
			{
				Name: "💰 Credits",
				Value: fmt.Sprintf("**Total Bet:** %s\n**Credits Won:** %s\n**Credits Lost:** %s\n**Net:** %s",
					p.Sprintf("%d", member.CreditsBet),
					p.Sprintf("%d", member.CreditsWon),
					p.Sprintf("%d", member.CreditsLost),
					formatNetCredits(netCredits, p)),
				Inline: false,
			},
		},
	}

	// Add last played field if the member has played before
	if !member.LastPlayed.IsZero() {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🕒 Last Played",
			Value:  fmt.Sprintf("<t:%d:R>", member.LastPlayed.Unix()),
			Inline: false,
		})
	}

	// Add footer with additional info
	if member.RoundsPlayed == 0 {
		embed.Description = "*No blackjack games played yet. Join a game to start tracking statistics!*"
	} else {
		avgHandsPerRound := float64(member.HandsPlayed) / float64(member.RoundsPlayed)
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Average %.1f hands per round", avgHandsPerRound),
		}
	}

	// Send ephemeral response
	resp := disgomsg.NewResponse(
		disgomsg.WithEmbeds([]*discordgo.MessageEmbed{embed}),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending stats response",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.String("targetUserID", targetUserID),
			slog.Any("error", err),
		)
	}
}

// formatNetCredits formats the net credits with appropriate color coding
func formatNetCredits(netCredits int, p *message.Printer) string {
	switch {
	case netCredits > 0:
		return fmt.Sprintf("**+%s** 📈", p.Sprintf("%d", netCredits))
	case netCredits < 0:
		return fmt.Sprintf("**%s** 📉", p.Sprintf("%d", netCredits))
	default:
		return "**0** ➖"
	}
}
