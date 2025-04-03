package leaderboard

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/olekukonko/tablewriter"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
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
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The system is shutting down."),
			disgomsg.WithInteraction(i.Interaction),
		)
		resp.SendEphemeral(s)
		return
	}

	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have permission to use this command."),
			disgomsg.WithInteraction(i.Interaction),
		)
		resp.SendEphemeral(s)
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
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The system is shutting down."),
			disgomsg.WithInteraction(i.Interaction),
		)
		resp.SendEphemeral(s)
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
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("Invalid command: "+options[0].Name),
			disgomsg.WithInteraction(i.Interaction),
		)
		resp.SendEphemeral(s)
	}
}

// currentLeaderboard returns the top ranked accounts for the current balance.
func currentLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	lb := getLeaderboard(i.GuildID)
	leaderboard := lb.getCurrentLeaderboard()
	sendLeaderboard(s, i, "Current Leaderboard", leaderboard)
}

// monthlyLeaderboard returns the top ranked accounts for the current months.
func monthlyLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	lb := getLeaderboard(i.GuildID)
	leaderboard := lb.getMonthlyLeaderboard()
	sendLeaderboard(s, i, "Monthly Leaderboard", leaderboard)
}

// lifetimeLeaderboard returns the top ranked accounts for the lifetime of the server.
func lifetimeLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	lb := getLeaderboard(i.GuildID)
	leaderboard := lb.getLifetimeLeaderboard()
	sendLeaderboard(s, i, "Lifetime Leaderboard", leaderboard)
}

// setLeaderboardChannel sets the server channel to which the monthly leaderboard is published.
func setLeaderboardChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	lb := getLeaderboard(i.GuildID)
	channelID := i.ApplicationCommandData().Options[0].Options[0].StringValue()
	lb.setChannel(channelID)

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Channel ID for the monthly leaderboard set to %s.", channelID)),
		disgomsg.WithInteraction(i.Interaction),
	)
	resp.Send(s)
}

// getLeaderboardInfo returns the leaderboard configuration for the server.
func getLeaderboardInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	lb := getLeaderboard(i.GuildID)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Channel ID for the monthly leaderboard is %s.", lb.ChannelID)),
		disgomsg.WithInteraction(i.Interaction),
	)
	resp.SendEphemeral(s)
}

// sendLeaderboard is a utility function that sends an economy leaderboard to Discord.
func sendLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate, title string, accounts []*bank.Account) {
	// Make sure the guild member's name is updated
	_ = guild.GetMember(i.GuildID, i.Member.User.ID).SetName(i.Member.User.Username, i.Member.DisplayName())

	p := message.NewPrinter(language.AmericanEnglish)
	embeds := formatAccounts(p, title, accounts)

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
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("**Current Rank**: %d\n**Monthly Rank**: %d\n**Lifetime Rank**: %d\n", currentRank, monthlyRank, lifetimeRank)),
		disgomsg.WithInteraction(i.Interaction),
	)
	resp.SendEphemeral(s)
}

// formatAccounts formats the leaderboard to be sent to a Discord server
func formatAccounts(p *message.Printer, title string, accounts []*bank.Account) []*discordgo.MessageEmbed {
	var tableBuffer strings.Builder
	table := tablewriter.NewWriter(&tableBuffer)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.SetHeader([]string{"#", "Name", "Balance"})

	// A bit of a hack, but good enough....
	for i, account := range accounts {
		member := guild.GetMember(accounts[0].GuildID, account.MemberID)
		var balance int
		switch title {
		case "Current Leaderboard":
			balance = account.CurrentBalance
		case "Monthly Leaderboard":
			balance = account.MonthlyBalance
		case "Lifetime Leaderboard":
			balance = account.LifetimeBalance
		default:
			balance = account.CurrentBalance
		}
		data := []string{strconv.Itoa(i + 1), member.Name, p.Sprintf("%d", balance)}
		table.Append(data)
	}
	table.Render()
	embeds := []*discordgo.MessageEmbed{
		{
			Type:  discordgo.EmbedTypeRich,
			Title: title,
			Fields: []*discordgo.MessageEmbedField{
				{
					Value: p.Sprintf("```\n%s```\n", tableBuffer.String()),
				},
			},
		},
	}

	return embeds
}
