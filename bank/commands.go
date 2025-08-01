package bank

import (
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"bank-admin": bankAdmin,
		"bank":       bank,
	}

	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "bank-admin",
			Description: "Commands used to interact with the economy for this server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "account",
					Description: "Sets the amount of credits for a given member.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The member or member ID.",
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
					Name:        "add",
					Description: "Adds the amount of credits to a given member's bank account.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The member or member ID.",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "amount",
							Description: "The amount to add to the account.",
							Required:    true,
						},
					},
				},
				{
					Name:        "balance",
					Description: "Set the default balance for the bank for the server.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "value",
							Description: "The the default balance for the bank for the server.",
							Required:    true,
						},
					},
				},
				{
					Name:        "name",
					Description: "Set the name of the bank for the server.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "value",
							Description: "The the name of the bank for the server.",
							Required:    true,
						},
					},
				},
				{
					Name:        "currency",
					Description: "Set the currency for the server.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "value",
							Description: "The currency to set for the server.",
							Required:    true,
						},
					},
				},
				{
					Name:        "info",
					Description: "Get information about the banking system configuration.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}

	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "bank",
			Description: "Commands used to interact with the economy for this server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "account",
					Description: "Bank account balance for the member.",
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

// bankAdmin routes the bankAdmin commands to the proper handers.
func bankAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The system is shutting down."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", i.GuildID),
				slog.String("error", err.Error()),
			)
		}
		return
	}

	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have permission to use this command."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", i.GuildID),
				slog.String("error", err.Error()),
			)
		}
		return
	}

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "balance":
		setDefaultBalance(s, i)
	case "name":
		setBankName(s, i)
	case "currency":
		setBankCurrency(s, i)
	case "account":
		setAccountBalance(s, i)
	case "add":
		addAccountBalance(s, i)
	case "info":
		getBankInfo(s, i)
	default:
		slog.Warn("unknown bank-admin command",
			slog.String("command", options[0].Name),
		)
	}
}

// bank routes the bank commands to the proper handlers.
func bank(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The system is shutting down."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", i.GuildID),
				slog.String("error", err.Error()),
			)
		}
		return
	}

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "account":
		account(s, i)
	default:
		slog.Warn("unknown bank command",
			slog.String("command", options[0].Name),
		)
	}
}

// account gets information about a member's bank account. By default, it uses the member who invoked the command.
// If an ID is provided, it will use that ID instead.
func account(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	memberID := i.Member.User.ID
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		if option.Name == "user" {
			user := option.UserValue(s)
			if user == nil {
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
			memberID = user.ID
			break
		}
	}

	account := GetAccount(i.GuildID, memberID)
	if account == nil {
		content := p.Sprintf("An account with the ID of %s does not exist.", memberID)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(content),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", i.GuildID),
				slog.String("error", err.Error()),
			)
		}
		return
	}

	content := p.Sprintf("**Current Balance**: %d\n**Monthly Balance**: %d\n**Lifetime Balance**: %d\n**Created**: %s\n",
		account.CurrentBalance,
		account.MonthlyBalance,
		account.LifetimeBalance,
		account.CreatedAt,
	)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(content),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("error", err.Error()),
		)
	}
}

// setAccountBalance sets the balance of the account for the member of the guild to the specified amount
func setAccountBalance(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	var memberID string
	var amount int
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		switch option.Name {
		case "user":
			user := option.UserValue(s)
			if user == nil {
				resp := disgomsg.NewResponse(
					disgomsg.WithContent("The account to set the balance for was not found. Please try again."),
				)
				if err := resp.SendEphemeral(s, i.Interaction); err != nil {
					slog.Error("error sending response",
						slog.String("guildID", i.GuildID),
						slog.String("error", err.Error()),
					)
				}
				return
			}
			memberID = user.ID
		case "amount":
			amount = int(option.IntValue())
		}
	}

	member, err := s.GuildMember(i.GuildID, memberID)
	if err != nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("An account with that ID does not exist."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", i.GuildID),
				slog.String("error", err.Error()),
			)
		}
		return
	}
	m := guild.GetMember(i.GuildID, memberID).SetName(member.User.Username, member.Nick, member.User.GlobalName)
	account := GetAccount(i.GuildID, memberID)

	if err := account.SetBalance(amount); err != nil {
		slog.Error("error setting bank account balance",
			slog.String("guildID", i.GuildID),
			slog.Int("amount", amount),
			slog.Any("error", err),
		)
	}

	slog.Debug("/bank-admin set account",
		slog.String("guildID", i.GuildID),
		slog.String("memberID", member.User.ID),
		slog.String("memberName", m.Name),
		slog.Int("balance", amount),
	)

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Account balance for %s was set to %d", m.Name, account.CurrentBalance)),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("error", err.Error()),
		)
	}
}

// addAccountBalance adds the amount to the balance of the account for the member of the guild
func addAccountBalance(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	var memberID string
	var amount int
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		switch option.Name {
		case "user":
			user := option.UserValue(s)
			if user == nil {
				resp := disgomsg.NewResponse(
					disgomsg.WithContent("The account to add credits to was not found. Please try again."),
				)
				if err := resp.SendEphemeral(s, i.Interaction); err != nil {
					slog.Error("error sending response",
						slog.String("guildID", i.GuildID),
						slog.String("error", err.Error()),
					)
				}
				return
			}
			memberID = user.ID
		case "amount":
			amount = int(option.IntValue())
		default:
			slog.Warn("unknown option in /bank-admin add command",
				slog.String("option", option.Name),
			)
		}
	}

	member, err := s.GuildMember(i.GuildID, memberID)
	if err != nil {
		slog.Error("error getting guild member",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", memberID),
			slog.Int("amount", amount),
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("An account with that ID does not exist."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", i.GuildID),
				slog.String("error", err.Error()),
			)
		}
		return
	}

	m := guild.GetMember(i.GuildID, memberID).SetName(member.User.Username, member.Nick, member.User.GlobalName)
	account := GetAccount(i.GuildID, memberID)

	if err := account.Deposit(amount); err != nil {
		slog.Error("error adding credits to the bank account balance",
			slog.String("guildID", i.GuildID),
			slog.Int("amount", amount),
			slog.Any("error", err),
		)
	}

	slog.Debug("/bank-admin add account",
		slog.String("guildID", i.GuildID),
		slog.String("memberID", memberID),
		slog.String("memberName", m.Name),
		slog.Int("amount", amount),
	)

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Account balance for %s was increased by %d", m.Name, amount)),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("error", err.Error()),
		)
	}
}

// setDefaultBalance sets the default balance for bank for the guild (server).
func setDefaultBalance(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var balance int
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		if option.Name == "value" {
			balance = int(option.IntValue())
			break
		}
	}

	p := message.NewPrinter(language.AmericanEnglish)

	bank := GetBank(i.GuildID)
	bank.SetDefaultBalance(balance)

	slog.Debug("/bank-admin balance",
		slog.String("guildID", i.GuildID),
		slog.Int("balance", balance),
	)

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Bank default balance was set to %d", bank.DefaultBalance)),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("error", err.Error()),
		)
	}
}

// setBankName sets the name of the bank for the guild (server).
func setBankName(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var name string
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		if option.Name == "value" {
			name = strings.TrimSpace(option.StringValue())
			break
		}
	}

	p := message.NewPrinter(language.AmericanEnglish)

	bank := GetBank(i.GuildID)
	bank.SetName(name)

	slog.Debug("bank-admin name",
		slog.String("guildID", i.GuildID),
		slog.String("name", name),
	)

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Bank name was set to %s", bank.Name)),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("error", err.Error()),
		)
	}
}

// setBankCurrency sets the name of the currency used by the bank for the guild (server).
func setBankCurrency(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var currency string
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		if option.Name == "value" {
			currency = strings.TrimSpace(option.StringValue())
			break
		}
	}

	p := message.NewPrinter(language.AmericanEnglish)

	bank := GetBank(i.GuildID)
	bank.SetCurrency(currency)

	slog.Debug("/bank-admin currency",
		slog.String("guildID", i.GuildID),
		slog.String("currency", currency),
	)

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Bank currency was set to %s", bank.Currency)),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("error", err.Error()),
		)
	}
}

// getBankInfo gets information about the bank for the guild (server).
func getBankInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	bank := GetBank(i.GuildID)

	content := p.Sprintf("**Bank Name**: %s\n**Currency**: %s\n**Default Balance**: %d\n",
		bank.Name,
		bank.Currency,
		bank.DefaultBalance,
	)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(content),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("error", err.Error()),
		)
	}
}
