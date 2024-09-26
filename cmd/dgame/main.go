package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rbrabson/dgame/database"
	"github.com/rbrabson/dgame/discord"
	log "github.com/sirupsen/logrus"
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
		log.SetLevel(log.DebugLevel)
	}
}

// Main Discord game bot
func main() {
	godotenv.Load()
	setLogLevel()

	db := database.NewClient()
	defer db.Close()

	bot := discord.NewBot()
	err := bot.Session.Open()
	if err != nil {
		log.WithField("error", err).Fatal("unable to create Discord bot")
	}
	defer bot.Session.Close()

	// Wait for the user to cancel the program
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	log.Info("Press Ctrl+C to exit")
	<-sc
}
