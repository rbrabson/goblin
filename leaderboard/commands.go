package leaderboard

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"

	"github.com/bwmarrin/discordgo"
	"github.com/olekukonko/tablewriter"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Type string

const (
	CurrentLeaderboard  Type = "Current Leaderboard"
	MonthlyLeaderboard  Type = "Monthly Leaderboard"
	LifetimeLeaderboard Type = "Lifetime Leaderboard"
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
					Description: "Sets the channel ID where the monthly leaderboard is published.",
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
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The member to return the leaderboard.",
							Required:    false,
						},
					},
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
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send the response",
				slog.Any("error", err),
			)
		}
		return
	}

	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have permission to use this command."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send the response",
				slog.Any("error", err),
			)
		}
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
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send the response",
				slog.Any("error", err),
			)
		}
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
			disgomsg.WithContent("Invalid command: " + options[0].Name),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send the response",
				slog.Any("error", err),
			)
		}
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

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Channel ID for the monthly leaderboard set to %s.", channelID)),
	)
	if err := resp.Send(s, i.Interaction); err != nil {
		slog.Error("failed to send the response",
			slog.Any("error", err),
		)
	}
}

// getLeaderboardInfo returns the leaderboard configuration for the server.
func getLeaderboardInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	lb := getLeaderboard(i.GuildID)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Channel ID for the monthly leaderboard is %s.", lb.ChannelID)),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("failed to send the response",
			slog.Any("error", err),
		)
	}
}

// sendLeaderboard is a utility function that sends an economy leaderboard to Discord.
func sendLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate, title Type, accounts []*bank.Account) {
	// Make sure the guild member's name is updated
	_ = guild.GetMember(i.GuildID, i.Member.User.ID).SetName(i.Member.User.Username, i.Member.Nick, i.Member.User.GlobalName)

	p := message.NewPrinter(language.AmericanEnglish)
	embeds := formatAccounts(p, string(title), accounts)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		slog.Error("failed to send the response",
			slog.Any("error", err),
		)
	}
}

// rank returns the rank of the member in the leaderboard.
func rank(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	memberID := i.Member.User.ID
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		if option.Name == "user" {
			member, err := guild.GetMemberByUser(s, i.GuildID, option.UserValue(s))
			if err != nil {
				resp := disgomsg.NewResponse(
					disgomsg.WithContent("The user you specified does not exist."),
				)
				if err := resp.SendEphemeral(s, i.Interaction); err != nil {
					slog.Error("error sending response",
						slog.String("guildID", i.GuildID),
						slog.String("error", err.Error()),
					)
				}
				return
			}
			memberID = member.MemberID
			break
		}
	}

	account := bank.GetAccount(i.GuildID, memberID)
	if account == nil {
		content := p.Sprintf("An account with the ID of %s does not exist.", memberID)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(content),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("failed to send the response",
				slog.Any("error", err),
			)
		}
		return
	}

	lb := getLeaderboard(i.GuildID)
	currentRank := getCurrentRanking(lb, account)
	monthlyRank := getMonthlyRanking(lb, account)
	lifetimeRank := getLifetimeRanking(lb, account)

	content := p.Sprintf("**Current Rank**: %d\n**Monthly Rank**: %d\n**Lifetime Rank**: %d\n", currentRank, monthlyRank, lifetimeRank)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(content),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("failed to send the response",
			slog.Any("error", err),
		)
	}
}

// formatAccounts formats the leaderboard to be sent to a Discord server
func formatAccounts(p *message.Printer, title string, accounts []*bank.Account) []*discordgo.MessageEmbed {
	var tableBuffer strings.Builder
	table := tablewriter.NewTable(&tableBuffer,
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Borders: tw.BorderNone,
			Symbols: tw.NewSymbols(tw.StyleASCII),
			Settings: tw.Settings{
				Separators: tw.Separators{BetweenRows: tw.Off, BetweenColumns: tw.Off},
				Lines:      tw.Lines{ShowHeaderLine: tw.Off},
			},
		})),
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Padding:    tw.CellPadding{Global: tw.Padding{Left: "", Right: "", Top: "", Bottom: ""}},
				Formatting: tw.CellFormatting{AutoWrap: tw.WrapNone}, // Wrap long content
				Alignment:  tw.CellAlignment{Global: tw.AlignLeft},   // Left-align rows
			},
			Header: tw.CellConfig{
				Padding:    tw.CellPadding{Global: tw.Padding{Left: "", Right: "", Top: "", Bottom: ""}},
				Formatting: tw.CellFormatting{AutoWrap: tw.WrapNone}, // Wrap long content
				Alignment:  tw.CellAlignment{Global: tw.AlignLeft},   // Left-align rows
			},
		}),
	)

	table.Header([]string{"#", "Name", "Balance"})

	// A bit of a hack, but good enough....
	for i, account := range accounts {
		member := guild.GetMember(accounts[0].GuildID, account.MemberID)
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
		data := []string{strconv.Itoa(i + 1), member.Name, p.Sprintf("%d", balance)}
		if err := table.Append(data); err != nil {
			slog.Error("failed to append data to the table",
				slog.Any("error", err),
			)
		}
	}
	if err := table.Render(); err != nil {
		slog.Error("failed to render the table",
			slog.Any("error", err),
		)
	}
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
