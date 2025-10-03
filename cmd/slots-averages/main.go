package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/database/mongo"
	"github.com/rbrabson/goblin/game/slots"
	"github.com/rbrabson/goblin/internal/log"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func main() {
	log.Initialize()

	err := godotenv.Load(".env")
	if err != nil {
		slog.LogAttrs(context.Background(), slog.LevelError,
			"unable to load .env_test file",
			slog.Any("error", err),
		)
	}

	guildID := os.Getenv("ACCOUNT_GUILD_ID")
	if guildID == "" {
		slog.Error("DISCORD_GUILD_ID environment variable not set")
		os.Exit(1)
	}

	db := mongo.NewDatabase()
	defer db.Close()
	slots.SetDB(db)

	averages, err := slots.GetPayoutAverages(guildID)
	if err != nil {
		slog.Error("error getting payout averages",
			slog.String("guildID", guildID),
			slog.Any("error", err),
		)
		return
	}

	p := message.NewPrinter(language.AmericanEnglish)

	totalGames := averages.TotalWins + averages.TotalLosses
	p.Printf("Total wins %d (%.2f%%)\n", averages.TotalWins, float64(averages.TotalWins)/float64(totalGames)*100)
	p.Printf("Total losses %d (%.2f%%)\n", averages.TotalLosses, float64(averages.TotalLosses)/float64(totalGames)*100)
	p.Printf("Total bet: %d\n", averages.TotalBet)
	p.Printf("Total won: %d\n", averages.TotalWon)
	p.Printf("Average wins: %.0f (%.2f%%)\n", averages.AverageTotalWins, averages.AverageWinPercentage)
	p.Printf("Average losses: %.0f (%.2f%%)\n", averages.AverageTotalLosses, averages.AverageLossPercentage)
	p.Printf("Average return: %.2f%% (%d won from %d bet)\n", averages.AverageReturns, averages.TotalWon, averages.TotalBet)
}
