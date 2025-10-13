package slots

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
	rslots "github.com/rbrabson/slots"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"slots": slots,
	}

	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "slots",
			Description: "Interacts with the slot machine.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "play",
					Description: "Play the slot machine.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "bet",
							Description: "The amount to bet on the slot machine.",
							Type:        discordgo.ApplicationCommandOptionInteger,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "500",
									Value: 500,
								},
								{
									Name:  "300",
									Value: 300,
								},
								{
									Name:  "100",
									Value: 100,
								},
							},
						},
					},
				},
				{
					Name:        "paytable",
					Description: "Get the pay table for the slot machine.",
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

// slots allows a user to play the slot machine.
func slots(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "play":
		playSlots(s, i)
	case "paytable":
		payTable(s, i)
	case "stats":
		showStats(s, i)
	}

}

// playSlots handles the `/slots play` command.
func playSlots(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	guildID := i.GuildID
	userID := i.Member.User.ID

	options := i.ApplicationCommandData().Options[0].Options
	bet := int(options[0].IntValue())

	slog.Debug("`/slots play` command",
		slog.String("guildID", guildID),
		slog.String("userID", userID),
		slog.Int("bet", bet),
	)

	config := GetConfig()
	member := GetMember(guildID, userID)
	if !member.IsInCooldown(config) {
		remaining := member.GetCooldownRemaining(config)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("You are on cooldown. Please wait %d seconds before playing again.", int(remaining.Seconds())+1)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", guildID),
				slog.String("userID", userID),
				slog.Any("error", err),
			)
		}
		return
	}

	account := bank.GetAccount(guildID, userID)
	if err := account.Withdraw(bet); err != nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have enough balance to play."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", guildID),
				slog.String("userID", userID),
				slog.Any("error", err),
			)
		}
		return
	}

	sm := GetSlotMachine()
	spinResult := sm.Spin(bet)

	member.AddResults(spinResult)

	if spinResult.Payout > 0 {
		if err := account.Deposit(spinResult.Payout); err != nil {
			slog.Error("error depositing slots winnings to account",
				slog.String("guildID", guildID),
				slog.String("userID", userID),
				slog.Int("payout", spinResult.Payout),
				slog.Any("error", err),
			)
		}
	}

	symbols := sm.symbols
	spinMsg := symbols["Blank"].Emoji
	for i, symbol := range spinResult.TopLine {
		if i != 0 {
			spinMsg += " | "
		}
		spinMsg += sm.symbols[symbol].Emoji
	}
	spinMsg += "\n" + symbols["Right Arrow"].Emoji
	for i, symbol := range spinResult.Payline {
		if i != 0 {
			spinMsg += " | "
		}
		spinMsg += sm.symbols[symbol].Emoji
	}
	spinMsg += "\n" + symbols["Blank"].Emoji
	for i, symbol := range spinResult.BottomLine {
		if i != 0 {
			spinMsg += " | "
		}
		spinMsg += symbols[symbol].Emoji
	}
	spinMsg += "\n"

	// Determine embed color based on win/loss
	var embedColor int
	var resultTitle string
	var resultDescription string

	if spinResult.Payout > 0 {
		embedColor = 0x00ff00 // Green for win
		resultTitle = "🎉 " + spinResult.Message
		resultDescription = p.Sprintf("You won **%d** coins!", spinResult.Payout)
	} else {
		embedColor = 0xff0000 // Red for loss
		resultTitle = "💸 No Win"
		resultDescription = "Better luck next time!"
	}

	// Create the embed
	embed := &discordgo.MessageEmbed{
		Title:       "🎰 Slot Machine 🎰",
		Description: p.Sprintf("<@%s> bet **%d** coins", userID, spinResult.Bet),
		Color:       embedColor,
		Fields: []*discordgo.MessageEmbedField{
			{
				Value:  spinMsg,
				Inline: false,
			},
			{
				Name:   resultTitle,
				Value:  resultDescription,
				Inline: false,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	resp := disgomsg.NewResponse(disgomsg.WithEmbeds([]*discordgo.MessageEmbed{embed}))
	resp.Send(s, i.Interaction)
}

// showStats handles the `/slots stats` command.
func showStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	memberID := i.Member.User.ID
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		if option.Name == "user" {
			var err error
			member, err := guild.GetMemberByUser(s, i.GuildID, option.UserValue(s))
			if err != nil {
				resp := disgomsg.NewResponse(
					disgomsg.WithContent("The user to get the account for was not found. Please try again."),
				)
				if err := resp.SendEphemeral(s, i.Interaction); err != nil {
					slog.Error("error sending response",
						slog.String("guildID", i.GuildID),
						slog.String("error", err.Error()),
					)
				}
				return
			}
			memberID = member.MemberID
		}
	}

	guildID := i.GuildID

	slog.Debug("`/slots stats` command",
		slog.String("guildID", guildID),
		slog.String("memberID", memberID),
	)

	member := GetMember(guildID, memberID)

	embed := &discordgo.MessageEmbed{
		Title:       "Slot Machine Stats",
		Description: p.Sprintf("Here are the stats for <@%s>:", memberID),
		Color:       0x5865F2, // Blue color
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Total Wins",
				Value:  p.Sprintf("%d", member.TotalWins),
				Inline: true,
			},
			{
				Name:   "Total Losses",
				Value:  p.Sprintf("%d", member.TotalLosses),
				Inline: true,
			},
			{
				Name:   "Winning Percentage",
				Value:  p.Sprintf("%.1f%%", (float64(member.TotalWins)/float64(member.TotalWins+member.TotalLosses))*100),
				Inline: true,
			},
			{
				Name:   "Total Bet",
				Value:  p.Sprintf("%d", member.TotalBet),
				Inline: true,
			},
			{
				Name:   "Total Winnings",
				Value:  p.Sprintf("%d", member.TotalWinnings),
				Inline: true,
			},
			{
				Name:   "Returns",
				Value:  p.Sprintf("%.1f%%", (float64(member.TotalWinnings)/float64(member.TotalBet))*100),
				Inline: true,
			},
			{
				Name:   "Current Winning Streak",
				Value:  p.Sprintf("%d", member.CurrentWinStreak),
				Inline: true,
			},
			{
				Name:   "Longest Winning Streak",
				Value:  p.Sprintf("%d", member.LongestWinStreak),
				Inline: true,
			},
			{
				Name:   "Max Win",
				Value:  p.Sprintf("%d", member.MaxWin),
				Inline: true,
			},
			{
				Name:   "Current Losing Streak",
				Value:  p.Sprintf("%d", member.CurrentLosingStreak),
				Inline: true,
			},
			{
				Name:   "Longest Losing Streak",
				Value:  p.Sprintf("%d", member.LongestLosingStreak),
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	resp := disgomsg.NewResponse(disgomsg.WithEmbeds([]*discordgo.MessageEmbed{embed}))
	resp.SendEphemeral(s, i.Interaction)
}

// payTable handles the `/slots paytable` command.
func payTable(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	guildID := i.GuildID
	payTable := GetPayoutTable()

	slog.Debug("`/slots paytable` command",
		slog.String("guildID", guildID),
	)

	embeds := []*discordgo.MessageEmbed{}
	if payTable != nil {
		embed := &discordgo.MessageEmbed{
			Title:       "Slot Machine Pay Table",
			Description: "Here are the possible winning combinations and their payouts.",
			Color:       0x00ff00, // Green color
			Fields:      make([]*discordgo.MessageEmbedField, 0, len(payTable)),
		}

		sm := GetSlotMachine()
		twoConsecTroops := false
		for _, payout := range payTable {
			payoutStr := strconv.FormatFloat(payout.Payout, 'f', -1, 64)
			betPayouts := p.Sprintf("Payout %s:%d\n", payoutStr, payout.Bet)

			name := getPayoutDisplayMessage(payout.Win, sm.symbols)
			if name == "two consecutive troops" {
				if twoConsecTroops {
					continue
				}
				twoConsecTroops = true
			}

			field := &discordgo.MessageEmbedField{
				Name:   getPayoutDisplayMessage(payout.Win, sm.symbols),
				Value:  betPayouts,
				Inline: false,
			}
			embed.Fields = append(embed.Fields, field)
		}

		embeds = append(embeds, embed)
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithContent("Pay table:"),
		disgomsg.WithEmbeds(embeds),
	)

	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", guildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
	}
}

// getPayoutDisplayMessage returns a user-friendly message for the payout symbols.
func getPayoutDisplayMessage(spin []string, symbolTable SymbolTable) string {
	if len(spin) == 1 {
		return spin[0]
	}

	symbols := make([]string, 0, len(spin))
	for _, symbol := range spin {
		if lookup, ok := symbolTable[symbol]; ok {
			symbols = append(symbols, lookup.Emoji)
			continue
		}

		switch symbol {
		case rslots.Any:
			symbols = append(symbols, "any symbol")
		case rslots.AnySeven:
			symbols = append(symbols, "hero")
		case rslots.AnyBar:
			symbols = append(symbols, "basic troop")
		case rslots.AnyRed:
			aq := symbolTable["red 7"]
			archer := symbolTable["1 bar"]
			symbols = append(symbols, fmt.Sprintf("%s/%s", aq.Emoji, archer.Emoji))
		case rslots.AnyWhite:
			gw := symbolTable["white 7"]
			wizard := symbolTable["2 bar"]
			symbols = append(symbols, fmt.Sprintf("%s/%s", gw.Emoji, wizard.Emoji))
		case rslots.AnyBlue:
			bk := symbolTable["blue 7"]
			barbarian := symbolTable["3 bar"]
			symbols = append(symbols, fmt.Sprintf("%s/%s", bk.Emoji, barbarian.Emoji))
		case rslots.MatchingNonBlank:
			symbols = append(symbols, "any troop")
		default:
			spell := symbolTable["blank"]
			symbols = append(symbols, spell.Emoji)
		}
	}
	display := strings.Join(symbols, " | ")
	switch display {
	case "any troop | any troop | any symbol", "any symbol | any troop | any troop":
		display = "two consecutive troops"
	}
	return display
}
