package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/game/heist"
	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/leaderboard"
	"github.com/rbrabson/goblin/payday"
	"github.com/rbrabson/goblin/server"
	log "github.com/sirupsen/logrus"
)

var (
	BotName  string = "Goblin"
	Version  string = "dev"
	Revision string = "test"
)

// setLogLevel sets the logging level. If the LOG_LEVEL environment variable isn't set or the value
// isn't recognized, logging defaults to the `debug` level
func setLogLevel() {
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "panic":
		log.SetLevel(log.PanicLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

// Main Discord game bot
func main() {
	err := godotenv.Load(".env_test")
	if err != nil {
		log.WithField("error", err).Warn("unable to load .env_test file")
	}
	setLogLevel()

	// Start the plugins
	bank.Start()
	heist.Start()
	guild.Start()
	leaderboard.Start()
	payday.Start()
	server.Start()

	bot := discord.NewBot(BotName, Version, Revision)
	err = bot.Session.Open()
	if err != nil {
		log.WithField("error", err).Fatal("unable to create Discord bot")
	}
	defer bot.Session.Close()

	// Wait for the user to cancel the program
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	log.Info("Press Ctrl+C to exit")
	<-sc
}
