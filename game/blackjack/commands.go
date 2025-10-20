package blackjack

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/discord"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"blackjack": blackjack,
	}

	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "blackjack",
			Description: "Interacts with the blackjack table.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "play",
					Description: "Play the blackjack game.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "stats",
					Description: "Shows a user's stats.",
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

// blackjack handles the /blackjack command and its subcommands.
func blackjack(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The system is shutting down."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", i.GuildID),
				slog.String("memberID", i.Member.User.ID),
				slog.Any("error", err),
			)
		}
		return
	}

	subCommand := i.ApplicationCommandData().Options[0].Name

	switch subCommand {
	case "play":
		playBlackjack(s, i)
	case "stats":
		showStats(s, i)
	default:
		// Unknown subcommand
	}
}

// playBlackjack handles the /blackjack/play command.
func playBlackjack(s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp := disgomsg.NewResponse(
		disgomsg.WithContent("Not Implemented Yet."),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
	}
}

// showStats handles the /blackjack/stats command.
func showStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp := disgomsg.NewResponse(
		disgomsg.WithContent("Not Implemented Yet."),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("error sending response",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
	}
}
