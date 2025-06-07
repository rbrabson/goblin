package leaderboard

import (
	"log/slog"
	"slices"
	"time"

	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/internal/disctime"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// A Leaderboard is used to send a monthly leaderboard to the Discord server for each guild.
type Leaderboard struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID    string             `json:"guild_id" bson:"guild_id"`
	ChannelID  string             `json:"channel_id" bson:"channel_id"`
	LastSeason time.Time          `json:"last_season" bson:"last_season"`
}

// newLeaderboard creates a new leaderboard for the given guildID and sets the last season to the current month.
func newLeaderboard(guildID string) *Leaderboard {
	lb := &Leaderboard{
		GuildID:    guildID,
		LastSeason: disctime.CurrentMonth(time.Now()),
	}
	if err := writeLeaderboard(lb); err != nil {
		slog.Error("Error writing leaderboard",
			slog.String("guild", guildID),
			slog.Any("error", err),
		)
	}

	return lb
}

// getLeaderboards returns all the leaderboards for all guilds known to the bot.
func getLeaderboards() []*Leaderboard {
	var leaderboards []*Leaderboard
	err := db.FindMany(LEADERBOARD_COLLECTION, bson.D{}, &leaderboards, bson.D{}, 0)
	if err != nil {
		slog.Error("unable to get leaderboards",
			slog.Any("error", err),
		)

		return nil
	}
	slog.Debug("leaderboards",
		slog.Int("count", len(leaderboards)),
	)
	return leaderboards
}

// getLeaderboard returns the leaderbord for the given guild
func getLeaderboard(guildID string) *Leaderboard {
	lb := readLeaderboard(guildID)
	if lb == nil {
		lb = newLeaderboard(guildID)
	}

	return lb
}

// setChannel sets the channel ID for the leaderboard to publish the monthly leaderboard.
func (lb *Leaderboard) setChannel(channelID string) {
	lb.ChannelID = channelID
	if err := writeLeaderboard(lb); err != nil {
		slog.Error("error writing leaderboard",
			slog.String("guild", lb.GuildID),
			slog.Any("error", err),
		)
	}
}

// GetCurrentRanking returns the global rankings based on the current balance.
func (lb *Leaderboard) getCurrentLeaderboard() []*bank.Account {
	filter := bson.D{{Key: "guild_id", Value: lb.GuildID}}
	sort := bson.D{{Key: "current_balance", Value: -1}, {Key: "_id", Value: 1}}
	limit := int64(10)

	accounts := bank.GetAccounts(lb.GuildID, filter, sort, limit)
	slices.SortFunc(accounts, func(a, b *bank.Account) int {
		switch {
		case a.CurrentBalance > b.CurrentBalance:
			return -1
		case a.CurrentBalance < b.CurrentBalance:
			return 1
		case a.MemberID < b.MemberID:
			return -1
		default:
			return 1
		}
	})

	return accounts
}

// getMonthlyLeaderboard returns the global rankings based on the monthly balance.
func (lb *Leaderboard) getMonthlyLeaderboard() []*bank.Account {
	filter := bson.D{{Key: "guild_id", Value: lb.GuildID}}
	sort := bson.D{{Key: "monthly_balance", Value: -1}, {Key: "_id", Value: 1}}
	limit := int64(10)

	accounts := bank.GetAccounts(lb.GuildID, filter, sort, limit)
	slices.SortFunc(accounts, func(a, b *bank.Account) int {
		switch {
		case a.MonthlyBalance > b.MonthlyBalance:
			return -1
		case a.MonthlyBalance < b.MonthlyBalance:
			return 1
		case a.MemberID < b.MemberID:
			return -1
		default:
			return 1
		}
	})

	return accounts
}

// getLifetimeLeaderboard returns the global rankings based on the monthly balance.
func (lb *Leaderboard) getLifetimeLeaderboard() []*bank.Account {
	filter := bson.D{{Key: "guild_id", Value: lb.GuildID}}
	sort := bson.D{{Key: "lifetime_balance", Value: -1}, {Key: "_id", Value: 1}}
	limit := int64(10)

	accounts := bank.GetAccounts(lb.GuildID, filter, sort, limit)
	slices.SortFunc(accounts, func(a, b *bank.Account) int {
		switch {
		case a.LifetimeBalance > b.LifetimeBalance:
			return -1
		case a.LifetimeBalance < b.LifetimeBalance:
			return 1
		case a.MemberID < b.MemberID:
			return -1
		default:
			return 1
		}
	})

	return accounts
}

// getMonthlyRanking returns the monthly global ranking on the server for a given player.
func getCurrentRanking(lb *Leaderboard, account *bank.Account) int {
	filter := bson.D{
		{Key: "guild_id", Value: lb.GuildID},
		{Key: "current_balance", Value: bson.D{{Key: "$gt", Value: account.CurrentBalance}}},
	}
	rank, _ := db.Count(bank.ACCOUNT_COLLECTION, filter)
	rank++
	slog.Debug("current ranking",
		slog.String("guildID", lb.GuildID),
		slog.String("account", account.MemberID),
		slog.Int("rank", rank),
	)

	return rank
}

// getMonthlyRanking returns the monthly global ranking on the server for a given player.
func getMonthlyRanking(lb *Leaderboard, account *bank.Account) int {
	filter := bson.D{
		{Key: "guild_id", Value: lb.GuildID},
		{Key: "monthly_balance", Value: bson.D{{Key: "$gt", Value: account.MonthlyBalance}}},
	}

	rank, _ := db.Count(bank.ACCOUNT_COLLECTION, filter)
	rank++
	slog.Debug("monthly ranking",
		slog.String("guildID", lb.GuildID),
		slog.String("account", account.MemberID),
		slog.Int("rank", rank),
	)

	return rank
}

// getLifetimeRanking returns the lifetime global ranking on the server for a given player.
func getLifetimeRanking(lb *Leaderboard, account *bank.Account) int {
	filter := bson.D{
		{Key: "guild_id", Value: lb.GuildID},
		{Key: "lifetime_balance", Value: bson.D{{Key: "$gt", Value: account.LifetimeBalance}}},
	}

	rank, _ := db.Count(bank.ACCOUNT_COLLECTION, filter)
	rank++
	slog.Debug("lifetime ranking",
		slog.String("guildID", lb.GuildID),
		slog.String("account", account.MemberID),
		slog.Int("rank", rank),
	)

	return rank
}

// sendMonthlyLeaderboard publishes the monthly leaderboard to the bank channel.
func sendMonthlyLeaderboard(lb *Leaderboard) error {
	// Get the top 10 accounts for this month
	accounts := lb.getMonthlyLeaderboard()
	leaderboardSize := min(10, len(accounts))
	accounts = accounts[:leaderboardSize]

	firstOfMonth := disctime.PreviousMonth(time.Now())
	year, month, _ := firstOfMonth.Date()
	if lb.ChannelID != "" {
		p := message.NewPrinter(language.AmericanEnglish)
		embeds := formatAccounts(p, fmt.Sprintf("%s %d Top 10", month, year), accounts)
		_, err := bot.Session.ChannelMessageSendComplex(lb.ChannelID, &discordgo.MessageSend{
			Embeds: embeds,
		})
		if err != nil {
			slog.Error("unable to send montly leaderboard",
				slog.String("guildID", lb.GuildID),
				slog.String("channelID", lb.ChannelID),
				slog.Any("error", err),
			)
			return err
		}
	} else {
		slog.Warn("no leaderboard channel set for server",
			slog.String("guildID", lb.GuildID),
			slog.String("channelID", lb.ChannelID),
		)
	}
	for _, account := range accounts {
		slog.Debug("sent monthly leaderboard",
			slog.String("guildID", lb.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Int("monthlyBalance", account.MonthlyBalance),
		)
	}
	slog.Info("sent monthly leaderboard",
		slog.String("guildID", lb.GuildID),
		slog.String("channelID", lb.ChannelID),
		slog.Int("leaderboardSize", leaderboardSize),
	)
	return nil
}

// publishMonthlyLeaderboard sends the monthly leaderboard to each guild.
func sendAllMonthlyLeaderboards() {
	// Get the last season for the banks, defaulting to the current time if there are no banks.
	// This handles the off-chance that the server crashed and a new month starts before the
	// server is restarted.
	lastSeason := time.Now()
	leaderboards := getLeaderboards()
	for _, lb := range leaderboards {
		if lb.LastSeason.Before(lastSeason) {
			lastSeason = lb.LastSeason
		}
	}

	for {
		nextMonth := disctime.NextMonth(lastSeason)
		time.Sleep(time.Until(nextMonth))
		lastSeason = disctime.CurrentMonth(lastSeason)

		leaderboards := getLeaderboards()
		for _, lb := range leaderboards {
			err := sendMonthlyLeaderboard(lb)
			if err != nil {
				slog.Error("unable to send monthly leaderboard",
					slog.String("guildID", lb.GuildID),
					slog.String("channelID", lb.ChannelID),
					slog.Any("error", err),
				)
			}
			lb.LastSeason = disctime.NextMonth(lastSeason)
			if err := writeLeaderboard(lb); err != nil {
				slog.Error("unable to write leaderboard to database",
					slog.String("guildID", lb.GuildID),
					slog.String("channelID", lb.ChannelID),
					slog.Any("error", err),
				)
			}
		}
		lastSeason = disctime.NextMonth(time.Now())

		bank.ResetMonthlyBalances()
	}
}

// String returns a string representation of the Leaderboard.
func (lb *Leaderboard) String() string {
	return fmt.Sprintf("Leaderboard{ID=%s, GuildID=%s, ChannelID=%s, LastSeason=%s}",
		lb.ID.Hex(),
		lb.GuildID,
		lb.ChannelID,
		lb.LastSeason,
	)
}
