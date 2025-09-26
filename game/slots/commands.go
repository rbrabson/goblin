package slots

import (
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/discord"
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
									Name:  "100",
									Value: 100,
								},
								{
									Name:  "200",
									Value: 200,
								},
								{
									Name:  "300",
									Value: 300,
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

	slog.Debug("play command",
		slog.String("guildID", guildID),
		slog.String("userID", userID),
		slog.Int("bet", bet),
	)

	account := bank.GetAccount(guildID, userID)
	if err := account.Withdraw(bet); err != nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have enough balance to play."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.Any("error", err),
			)
		}
		return
	}

	sm := NewSlotMachine(guildID)
	spinResult := sm.Spin(bet)

	member := GetMemberKey(guildID, userID)
	member.AddResults(spinResult)

	if spinResult.Payout > 0 {
		if err := account.Deposit(spinResult.Payout); err != nil {
			slog.Error("error depositing winnings to account",
				slog.String("guildID", guildID),
				slog.String("userID", userID),
				slog.Int("payout", spinResult.Payout),
				slog.Any("error", err),
			)
		}
	}

	symbols := sm.Symbols.Symbols
	spinMsg := symbols["Blank"].Emoji
	for _, symbol := range spinResult.NextLine {
		spinMsg += symbol.Emoji
	}
	spinMsg += "\n" + symbols["Right Arrow"].Emoji
	for _, symbol := range spinResult.Payline {
		spinMsg += symbol.Emoji
	}
	spinMsg += "\n" + symbols["Blank"].Emoji
	for _, symbol := range spinResult.PreviousLine {
		spinMsg += symbol.Emoji
	}
	spinMsg += "\n"

	// Determine embed color based on win/loss
	var embedColor int
	var resultTitle string
	var resultDescription string

	if spinResult.Payout > 0 {
		embedColor = 0x00ff00 // Green for win
		resultTitle = "🎉 Winner!"
		resultDescription = p.Sprintf("You won **%d** coins!", spinResult.Payout)
	} else {
		embedColor = 0xff0000 // Red for loss
		resultTitle = "💸 No Win"
		resultDescription = "Better luck next time!"
	}

	// Create the embed
	embed := &discordgo.MessageEmbed{
		Title:       "Slot Machine",
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

// payTable handles the `/slots paytable` command.
func payTable(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	guildID := i.GuildID
	payTable := GetPayoutTable(guildID)

	slog.Debug("paytable command",
		slog.String("guildID", guildID),
		slog.Any("payTable", payTable),
	)

	embeds := []*discordgo.MessageEmbed{}
	if payTable != nil {
		embed := &discordgo.MessageEmbed{
			Title:       "Slot Machine Pay Table",
			Description: "Here are the possible winning combinations and their payouts.",
			Color:       0x00ff00, // Green color
			Fields:      make([]*discordgo.MessageEmbedField, 0, len(payTable.Payouts)),
		}

		for _, payout := range payTable.Payouts {
			winCombination := ""
			if len(payout.Win) == 1 {
				winCombination = payout.Win[0]
			} else {
				winCombination = "[" + payout.Win[0]
				for _, win := range payout.Win[1:] {
					winCombination += ", " + win
				}
				winCombination += "]"
			}

			betPayouts := p.Sprintf("Payout %d:%d\n", payout.Payout, payout.Bet)

			field := &discordgo.MessageEmbedField{
				Name:   winCombination,
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
			slog.Any("error", err),
		)
	}
}

// showStats handles the `/slots stats` command.
func showStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	guildID := i.GuildID
	userID := i.Member.User.ID

	slog.Debug("stats command",
		slog.String("guildID", guildID),
		slog.String("userID", userID),
	)

	member := GetMemberKey(guildID, userID)

	embed := &discordgo.MessageEmbed{
		Title:       "Slot Machine Stats",
		Description: p.Sprintf("Here are the stats for <@%s>:", userID),
		Color:       0x00ff00, // Green color
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
				Name:   "Current Win Streak",
				Value:  p.Sprintf("%d", member.CurrentWinStreak),
				Inline: true,
			},
			{
				Name:   "Longest Win Streak",
				Value:  p.Sprintf("%d", member.LongestWinStreak),
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	resp := disgomsg.NewResponse(disgomsg.WithEmbeds([]*discordgo.MessageEmbed{embed}))
	resp.SendEphemeral(s, i.Interaction)
}
