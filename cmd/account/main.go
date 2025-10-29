package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/database/mongo"
	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/internal/log"
	"github.com/rbrabson/goblin/stats"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	PlayerStatsCollection = "player_stats"
)

var (
	db *mongo.MongoDB
)

type Players struct {
	Stats        *stats.PlayerStats
	Name         string
	LengthPlayed time.Duration
}

func getPlayerStats(guildID string, game string) []*stats.PlayerStats {
	var ps []*stats.PlayerStats
	filter := bson.M{"guild_id": guildID, "game": game}
	err := db.FindMany(PlayerStatsCollection, filter, &ps, nil, 0)
	if err != nil {
		return nil
	}
	return ps
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		slog.LogAttrs(context.Background(), slog.LevelError,
			"unable to load .env file",
			slog.Any("error", err),
		)
	}

	log.Initialize()

	guildID := os.Getenv("ACCOUNT_GUILD_ID")
	if guildID == "" {
		slog.Error("DISCORD_GUILD_ID environment variable not set")
		os.Exit(1)
	}

	db = mongo.NewDatabase()
	defer db.Close()
	guild.SetDB(db)

	// Get all players who have

	heistStats := getPlayerStats(guildID, "heist")
	raceStats := getPlayerStats(guildID, "race")

	ps := make(map[string]*stats.PlayerStats)
	for _, p := range heistStats {
		ps[p.MemberID] = p
	}
	for _, p := range raceStats {
		if existing, ok := ps[p.MemberID]; ok {
			existing.FirstPlayed = minTime(existing.FirstPlayed, p.FirstPlayed)
			existing.LastPlayed = maxTime(existing.LastPlayed, p.LastPlayed)
			existing.NumberOfTimesPlayed += p.NumberOfTimesPlayed
		} else {
			ps[p.MemberID] = p
		}
	}

	players := make([]*Players, 0, len(ps))
	for _, p := range ps {
		member := guild.GetMember(guildID, p.MemberID)
		player := &Players{
			Stats:        p,
			Name:         member.Name,
			LengthPlayed: p.LastPlayed.Sub(p.FirstPlayed),
		}
		players = append(players, player)
	}

	slices.SortFunc(players, func(a, b *Players) int {
		if a.LengthPlayed != b.LengthPlayed {
			return int(b.LengthPlayed - a.LengthPlayed)
		}
		if a.Stats.NumberOfTimesPlayed != b.Stats.NumberOfTimesPlayed {
			return b.Stats.NumberOfTimesPlayed - a.Stats.NumberOfTimesPlayed
		}
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})

	fmt.Printf("%-20s,%3s,%10s,%10s,%10s\n", "Name", "Times Played", "First Played", "Last Played", "Days Played")
	for _, p := range players {
		daysPlayed := int(p.LengthPlayed.Truncate(time.Hour).Hours()/24) + 1
		fmt.Printf("%-20s,%3d,%10s,%10s,%d\n", p.Name, p.Stats.NumberOfTimesPlayed, p.Stats.FirstPlayed.Format("2006-01-02"), p.Stats.LastPlayed.Format("2006-01-02"), daysPlayed)
	}
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
