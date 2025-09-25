package slots

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
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

	var resp *disgomsg.Response

	lookupTable := GetLookupTable(guildID)
	if lookupTable == nil {
		resp = disgomsg.NewResponse(
			disgomsg.WithContent("No slot machine configured for this server."),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}
	payoutTable := GetPayoutTable(guildID)
	if payoutTable == nil {
		resp = disgomsg.NewResponse(
			disgomsg.WithContent("No payout table configured for this server."),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	symbols := GetSymbols(guildID)

	spinResult := lookupTable.Spin()
	spin := spinResult.Spins[spinResult.WinIndex]
	payout := payoutTable.GetPayoutAmount(bet, spin)

	betMsg := "You played the slot machine!\n"

	var spinMsg string
	for idx, spin := range spinResult.Spins {
		if idx < len(spinResult.Spins)-3 {
			continue
		}
		if idx == spinResult.WinIndex {
			spinMsg += symbols.Symbols["Right Arrow"].Emoji
		} else {
			spinMsg += symbols.Symbols["Blank"].Emoji
		}
		for _, symbol := range spin {
			spinMsg += symbol.Emoji
		}
		spinMsg += "\n"
	}

	var payoutMsg string
	if payout > 0 {
		payoutMsg = p.Sprintf("<@%s> bet %d and won %d!", i.Member.User.ID, bet, payout)
	} else {
		payoutMsg = p.Sprintf("<@%s> lost their bet of %d.", i.Member.User.ID, bet)
	}

	disgomsg.NewResponse(disgomsg.WithContent(betMsg+spinMsg+payoutMsg)).Send(s, i.Interaction)
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
			Description: "Here are the possible winning combinations and their payouts based on your bet amount.",
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

			betPayouts := ""
			for bet, amount := range payout.Bet {
				betPayouts += p.Sprintf("%d: %d\n", bet, amount)
			}

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
