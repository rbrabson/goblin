package bank

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/dgame/guild"
	"github.com/rbrabson/dgame/msg"
	log "github.com/sirupsen/logrus"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"bank": bank,
	}

	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "bank",
			Description: "Commands used to interact with the economy for this server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "account",
					Description: "Gets the bank account information for the given member.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "id",
							Description: "The member ID.",
							Required:    true,
						},
					},
				},
				{
					Name:        "set",
					Description: "Sets the amount of credits for a given member.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "id",
							Description: "The member ID.",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "amount",
							Description: "The amount to set the account to.",
							Required:    true,
						},
					},
				},
				{
					Name:        "channel",
					Description: "Sets the channel ID where the monthly leaderboard is published at the end of the month.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "id",
							Description: "The channel ID.",
							Required:    true,
						},
					},
				},
			},
		},
	}
	memberCommands = []*discordgo.ApplicationCommand{}
)

// bank routes the bank commands to the proper handers.
func bank(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> bank.bank")
	defer log.Trace("<-- bank.bank")

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "account":
		getAccount(s, i)
	case "channel":
		setChannel(s, i)
	case "set":
		setBalance(s, i)
	}
}

func setChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> setLeaderboardChannel")
	defer log.Trace("<-- setLeaderboardChannel")

	p := msg.GetPrinter(i)

	guild := guild.GetGuild(i.GuildID)
	bank := GetBank(guild)
	channelID := i.ApplicationCommandData().Options[0].Options[0].StringValue()
	bank.ChannelID = channelID

	bank.Write()

	resp := p.Sprintf("Channel ID for the monthly leaderboard set to %s.", bank.ChannelID)
	msg.SendResponse(s, i, resp)
}

// setBalance sets the balance of the account for the member of the guild to the specified amount
func setBalance(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> setBalance")
	defer log.Trace("<-- setBalance")

	var id string
	var amount uint
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		switch option.Name {
		case "id":
			id = strings.TrimSpace(option.StringValue())
		case "amount":
			amount = uint(option.IntValue())
		}
	}

	p := msg.GetPrinter(i)

	member, err := s.GuildMember(i.GuildID, id)
	if err != nil {
		resp := p.Sprintf("An account with ID `%s` is not a member of this server", id)
		msg.SendEphemeralResponse(s, i, resp)
		return
	}

	g := guild.GetGuild(i.GuildID)
	m := guild.GetMember(g, member.User.ID).SetName(i.Member.DisplayName())
	bank := GetBank(g)
	account := GetAccount(bank, id)

	account.SetBalance(amount)

	log.WithFields(log.Fields{
		"guild":   g.ID,
		"member":  m.ID,
		"balance": amount,
	}).Debug("/bank set")

	account.Write()

	resp := p.Sprintf("Account balance for %s was set to %d", m.Name, account.Balance)
	msg.SendResponse(s, i, resp)
}

// getAccount returns information about a bank getAccount for the specified member.
func getAccount(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> accountInfo")
	defer log.Trace("<-- accountInfo")

	p := msg.GetPrinter(i)

	g := guild.GetGuild(i.GuildID)
	m := guild.GetMember(g, i.Member.User.ID).SetName(i.Member.DisplayName())
	bank := GetBank(g)
	account := GetAccount(bank, m.ID)

	resp := p.Sprintf("**Name**: %s\n**ID**: %s\n**Balance**: %d\n**Lifetime**: %d", m.Name, m.ID, account.Balance, account.LifetimeDeposits)
	msg.SendEphemeralResponse(s, i, resp)
}
