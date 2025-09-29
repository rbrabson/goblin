package payday

import (
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/internal/format"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"payday": payday,
	}

	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "payday",
			Description: "Deposits your daily check into your bank account.",
			// Options: []*discordgo.ApplicationCommandOption{
			// 	{
			// 		Name:        "stats",
			// 		Description: "View your payday statistics.",
			// 		Type:        discordgo.ApplicationCommandOptionSubCommand,
			// 		Required:    false,
			// 	},
			// },
		},
	}
)

// payday handles the `/payday` command.
func payday(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	stats := false
	options := i.ApplicationCommandData().Options
	for _, option := range options {
		if option.Name == "stats" {
			stats = true
		}
	}

	if stats {
		showStats(s, i)
	} else {
		processPayday(s, i)
	}
}

// processPayday processes the `/payday` command without the `stats` option.
func processPayday(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	payday := GetPayday(i.GuildID)
	paydayAccount := payday.GetAccount(i.Member.User.ID)

	if paydayAccount.getNextPayday().After(time.Now()) {
		remainingTime := time.Until(paydayAccount.NextPayday)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(p.Sprintf("You can't get another payday yet. You need to wait %s.", format.Duration(remainingTime))),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.Any("error", err),
			)
		}
		return
	}

	paydayAmount := paydayAccount.getPayAmmount()

	account := bank.GetAccount(i.GuildID, i.Member.User.ID)
	if err := account.Deposit(paydayAmount); err != nil {
		slog.Error("error depositing data in the account",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
	}

	paydayAccount.setNextPayday(payday.PaydayFrequency)

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("You deposited your check of %d into your bank account. You now have %d credits.", paydayAmount, account.CurrentBalance)),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.Any("error", err),
		)
	}
}

// showStats handles the `/payday stats` command.
func showStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	payday := GetPayday(i.GuildID)
	paydayAccount := payday.GetAccount(i.Member.User.ID)
	currentStreak := paydayAccount.CurrentStreak
	maxStreak := paydayAccount.MaxStreak
	pay := paydayAccount.getPayAmmount()
	nextPayday := paydayAccount.getNextPayday().Format(time.DateTime)

	embeds := []*discordgo.MessageEmbed{
		{
			Title:       "Payday Stats",
			Description: "Here are your payday statistics.",
			Color:       0x00ff00, // Green
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Payday Amount",
					Value:  p.Sprintf("%d", pay),
					Inline: true,
				},
				{
					Name:   "Current Streak",
					Value:  p.Sprintf("%d", currentStreak),
					Inline: true,
				},
				{
					Name:   "Max Streak",
					Value:  p.Sprintf("%d", maxStreak),
					Inline: true,
				},
				{
					Name:   "Next Payday",
					Value:  nextPayday,
					Inline: true,
				},
			},
		},
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithEmbeds(embeds),
	)

	guildID := i.GuildID
	memberID := i.Member.User.ID

	slog.Debug("`/payday stats` command",
		slog.String("guildID", guildID),
		slog.String("memberID", memberID),
	)

	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", guildID),
			slog.String("memberID", memberID),
			slog.Any("error", err),
		)
	}
}
