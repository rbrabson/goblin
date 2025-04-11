package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/game/heist"
	"github.com/rbrabson/goblin/game/race"
	"github.com/rbrabson/goblin/internal/logger"
	"github.com/rbrabson/goblin/leaderboard"
	"github.com/rbrabson/goblin/payday"
	"github.com/rbrabson/goblin/role"
	"github.com/rbrabson/goblin/shop"
)

var (
	BotName  string = "Goblin"
	Version  string = "dev"
	Revision string = "test"
)

// Main Discord game bot
func main() {
	sslog := logger.GetLogger()

	err := godotenv.Load(".env")
	if err != nil {
		sslog.Warn("unable to load .env_test file", slog.Any("error", err))
	}

	// Start the plugins
	bank.Start()
	heist.Start()
	leaderboard.Start()
	payday.Start()
	race.Start()
	role.Start()
	shop.Start()

	bot := discord.NewBot(BotName, Version, Revision)
	err = bot.Session.Open()
	if err != nil {
		sslog.Error("unable to create Discord bot", slog.Any("error", err))
		os.Exit(1)
	}
	defer bot.Session.Close()

	// Wait for the user to cancel the program
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	sslog.Info("Press Ctrl+C to exit")
	<-sc

	// Close down the bot's session to Discord
	err = bot.Session.Close()
	if err != nil {
		sslog.Error("failed to gracefully close the Discord session", slog.Any("error", err))
	}
}
