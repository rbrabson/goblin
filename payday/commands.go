package payday

import (
	"fmt"
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
		resp := disgomsg.Response{
			Content: "The system is shutting down.",
		}
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	p := message.NewPrinter(language.AmericanEnglish)
	payday := GetPayday(i.GuildID)
	paydayAccount := payday.GetAccount(i.Member.User.ID)

	if paydayAccount.getNextPayday().After(time.Now()) {
		remainingTime := time.Until(paydayAccount.NextPayday)
		resp := disgomsg.Response{
			Content: fmt.Sprintf("You can't get another payday yet. You need to wait %s.", format.Duration(remainingTime)),
		}
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	account := bank.GetAccount(i.GuildID, i.Member.User.ID)
	account.Deposit(payday.Amount)

	paydayAccount.setNextPayday(time.Now().Add(payday.PaydayFrequency))

	resp := disgomsg.Response{
		Content: p.Sprintf("You deposited your check of %d into your bank account. You now have %d credits.", payday.Amount, account.CurrentBalance),
	}
	resp.SendEphemeral(s, i.Interaction)
}
