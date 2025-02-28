package payday

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/internal/discmsg"
	"github.com/rbrabson/goblin/internal/format"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
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
	log.Trace("--> payday")
	defer log.Trace("<-- payday")

	p := discmsg.GetPrinter(language.AmericanEnglish)
	payday := GetPayday(i.GuildID)
	paydayAccount := payday.GetAccount(i.Member.User.ID)

	if paydayAccount.getNextPayday().After(time.Now()) {
		remainingTime := time.Until(paydayAccount.NextPayday)
		resp := p.Sprintf("You can't get another payday yet. You need to wait %s.", format.Duration(remainingTime))
		discmsg.SendEphemeralResponse(s, i, resp)
		return
	}

	account := bank.GetAccount(i.GuildID, i.Member.User.ID)
	account.Deposit(payday.Amount)

	paydayAccount.setNextPayday(time.Now().Add(payday.PaydayFrequency))

	resp := p.Sprintf("You deposited your check of %d into your bank account. You now have %d credits.", payday.Amount, account.CurrentBalance)
	discmsg.SendEphemeralResponse(s, i, resp)
}
