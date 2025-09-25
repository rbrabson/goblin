package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/game/heist"
	"github.com/rbrabson/goblin/game/race"
	"github.com/rbrabson/goblin/game/slots"
	"github.com/rbrabson/goblin/internal/log"
	"github.com/rbrabson/goblin/leaderboard"
	"github.com/rbrabson/goblin/payday"
	"github.com/rbrabson/goblin/role"
	"github.com/rbrabson/goblin/shop"
	"github.com/rbrabson/goblin/stats"
)

var (
	BotName  = "Goblin"
	Version  = "dev"
	Revision = "test"
)

// Main Discord game bot
func main() {
	log.Initialize()

	err := godotenv.Load(".env")
	if err != nil {
		slog.LogAttrs(context.Background(), slog.LevelError,
			"unable to load .env_test file",
			slog.Any("error", err),
		)
	}

	// Start the plugins
	// account.Start()
	bank.Start()
	heist.Start()
	leaderboard.Start()
	payday.Start()
	race.Start()
	role.Start()
	shop.Start()
	slots.Start()
	stats.Start()

	bot := discord.NewBot(BotName, Version, Revision)
	err = bot.Session.Open()
	if err != nil {
		slog.LogAttrs(context.Background(), slog.LevelError,
			"unable to create Discord bot",
			slog.Any("error", err),
		)
		os.Exit(1)
	}
	defer func(Session *discordgo.Session) {
		err := Session.Close()
		if err != nil {
			slog.Error("unable to close Discord session",
				slog.Any("error", err),
			)
		}
	}(bot.Session)

	// Wait for the user to cancel the program
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	slog.LogAttrs(context.Background(), slog.LevelInfo,
		"Press Ctrl+C to exit",
	)
	<-sc

	// Close down the bot's session to Discord
	err = bot.Session.Close()
	if err != nil {
		slog.LogAttrs(context.Background(), slog.LevelError,
			"failed to gracefully close the Discord session",
			slog.Any("error", err),
		)
	}
}
