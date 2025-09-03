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
		},
	}
)

// payday gives some credits to the player every 24 hours.
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

	account := bank.GetAccount(i.GuildID, i.Member.User.ID)
	if err := account.Deposit(payday.Amount); err != nil {
		slog.Error("error depositing data in the account",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
	}

	paydayAccount.setNextPayday(time.Now().Add(payday.PaydayFrequency))

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("You deposited your check of %d into your bank account. You now have %d credits.", payday.Amount, account.CurrentBalance)),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.Any("error", err),
		)
	}
}
