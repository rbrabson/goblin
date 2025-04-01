package leaderboard

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type LeaderboardType string

const (
	CurrentLeaderboard  LeaderboardType = "Current Leaderboard"
	MonthlyLeaderboard  LeaderboardType = "Monthly Leaderboard"
	LifetimeLeaderboard LeaderboardType = "Lifetime Leaderboard"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"lb-admin": leaderboardAdmin,
		"lb":       leaderboard,
	}

	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "lb-admin",
			Description: "Commands used to interact with the leaderboard for this server.",
			Options: []*discordgo.ApplicationCommandOption{
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
				{
					Name:        "info",
					Description: "Gets information about the leaderboard configuration.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}
	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "lb",
			Description: "Commands used to retrieve leaderboards on this server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "current",
					Description: "Gets the current economy leaderboard.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "monthly",
					Description: "Gets the monthly economy leaderboard.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "lifetime",
					Description: "Gets the lifetime economy leaderboard.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "rank",
					Description: "Gets the member rank for the leaderboards.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}
)

// leaderboardAdmin updates the leaderboardAdmin channel.
func leaderboardAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		resp := disgomsg.Response{
			Content: "The system is shutting down.",
		}
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := disgomsg.Response{
			Content: "You do not have permission to use this command.",
		}
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	options := i.ApplicationCommandData().Options
	if options[0].Name == "channel" {
		setLeaderboardChannel(s, i)
	} else if options[0].Name == "info" {
		getLeaderboardInfo(s, i)
	}
}

// leaderboard handles the leaderboard commands.
func leaderboard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		resp := disgomsg.Response{
			Content: "The system is shutting down.",
		}
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "current":
		currentLeaderboard(s, i)
	case "monthly":
		monthlyLeaderboard(s, i)
	case "lifetime":
		lifetimeLeaderboard(s, i)
	case "rank":
		rank(s, i)
	case "type":
		switch options[0].IntValue() {
		case 1:
			monthlyLeaderboard(s, i)
		case 2:
			currentLeaderboard(s, i)
		case 3:
			lifetimeLeaderboard(s, i)
		case 4:
			rank(s, i)
		}
	default:
		resp := disgomsg.Response{
			Content: "Invalid command: " + options[0].Name,
		}
		resp.SendEphemeral(s, i.Interaction)
	}
}

// currentLeaderboard returns the top ranked accounts for the current balance.
func currentLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	lb := getLeaderboard(i.GuildID)
	leaderboard := lb.getCurrentLeaderboard()
	sendLeaderboard(s, i, CurrentLeaderboard, leaderboard)
}

// monthlyLeaderboard returns the top ranked accounts for the current months.
func monthlyLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	lb := getLeaderboard(i.GuildID)
	leaderboard := lb.getMonthlyLeaderboard()
	sendLeaderboard(s, i, MonthlyLeaderboard, leaderboard)
}

// lifetimeLeaderboard returns the top ranked accounts for the lifetime of the server.
func lifetimeLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	lb := getLeaderboard(i.GuildID)
	leaderboard := lb.getLifetimeLeaderboard()
	sendLeaderboard(s, i, LifetimeLeaderboard, leaderboard)
}

// setLeaderboardChannel sets the server channel to which the monthly leaderboard is published.
func setLeaderboardChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	lb := getLeaderboard(i.GuildID)
	channelID := i.ApplicationCommandData().Options[0].Options[0].StringValue()
	lb.setChannel(channelID)

	resp := disgomsg.Response{
		Content: fmt.Sprintf("Channel ID for the monthly leaderboard set to %s.", lb.ChannelID),
	}
	resp.Send(s, i.Interaction)
}

// getLeaderboardInfo returns the leaderboard configuration for the server.
func getLeaderboardInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	lb := getLeaderboard(i.GuildID)
	resp := disgomsg.Response{
		Content: fmt.Sprintf("channel ID for the monthly leaderboard is %s.", lb.ChannelID),
	}
	resp.SendEphemeral(s, i.Interaction)
}

// sendLeaderboard is a utility function that sends an economy leaderboard to Discord.
func sendLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate, title LeaderboardType, accounts []*bank.Account) {
	// Make sure the guild member's name is updated
	_ = guild.GetMember(i.GuildID, i.Member.User.ID).SetName(i.Member.User.Username, i.Member.Nick, i.Member.User.GlobalName)

	p := message.NewPrinter(language.AmericanEnglish)
	embeds := formatAccounts(p, string(title), accounts)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

// rank returns the rank of the member in the leaderboard.
func rank(s *discordgo.Session, i *discordgo.InteractionCreate) {
	account := bank.GetAccount(i.GuildID, i.Member.User.ID)
	lb := getLeaderboard(i.GuildID)
	currentRank := getCurrentRanking(lb, account)
	monthlyRank := getMonthlyRanking(lb, account)
	lifetimeRank := getLifetimeRanking(lb, account)

	p := message.NewPrinter(language.AmericanEnglish)
	resp := disgomsg.Response{
		Content: p.Sprintf("**Current Rank**: %d\n**Monthly Rank**: %d\n**Lifetime Rank**: %d\n", currentRank, monthlyRank, lifetimeRank),
	}
	resp.SendEphemeral(s, i.Interaction)
}

// formatAccounts formats the leaderboard to be sent to a Discord server
func formatAccounts(p *message.Printer, title string, accounts []*bank.Account) []*discordgo.MessageEmbed {
	embeds := []*discordgo.MessageEmbed{
		{
			Type:   discordgo.EmbedTypeRich,
			Title:  title,
			Fields: make([]*discordgo.MessageEmbedField, 0, len(accounts)),
		},
	}
	embed := embeds[0]

	sb := strings.Builder{}
	for i, account := range accounts {
		sb.Reset()
		member := guild.GetMember(account.GuildID, account.MemberID)
		var balance int
		switch title {
		case string(CurrentLeaderboard):
			balance = account.CurrentBalance
		case string(MonthlyLeaderboard):
			balance = account.MonthlyBalance
		case string(LifetimeLeaderboard):
			balance = account.LifetimeBalance
		default:
			balance = account.MonthlyBalance
		}

		sb.WriteString(p.Sprintf("Rank %d", i+1))
		switch i {
		case 0:
			sb.WriteString(" 🥇")
		case 1:
			sb.WriteString(" 🥈")
		case 2:
			sb.WriteString(" 🥉")
		}
		sb.WriteString(p.Sprintf(" - %s", member.Name))
		name := sb.String()

		sb.Reset()
		sb.WriteString(p.Sprintf("Balance: %d\n", balance))
		sb.WriteString(p.Sprintf("Tag: <@%s>\n", member.MemberID))
		if member.UserName != "" {
			sb.WriteString(p.Sprintf("Username: %s\n", member.UserName))
		}
		value := sb.String()

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   name,
			Value:  value,
			Inline: false,
		})
	}

	return embeds
}
