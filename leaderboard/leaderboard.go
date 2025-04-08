package leaderboard

import (
	"slices"
	"time"

	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/internal/disctime"
	log "github.com/sirupsen/logrus"
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
	writeLeaderboard(lb)
	log.WithFields(log.Fields{"guildID": guildID, "leaderboard": lb}).Trace("new leaderboard")

	return lb
}

// getLeaderboards returns all the leaderboards for all guilds known to the bot.
func getLeaderboards() []*Leaderboard {
	var leaderboards []*Leaderboard
	err := db.FindMany(LEADERBOARD_COLLECTION, bson.D{}, &leaderboards, bson.D{}, 0)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("unable to get leaderboards")
		return nil
	}
	log.WithFields(log.Fields{"leaderboards": leaderboards}).Debug("leaderboards")
	return leaderboards
}

// getLeaderboard returns the leaderbord for the given guild
func getLeaderboard(guildID string) *Leaderboard {
	lb := readLeaderboard(guildID)
	if lb == nil {
		lb = newLeaderboard(guildID)
	}
	log.WithFields(log.Fields{"guildID": guildID, "leaderboard": lb}).Trace("leaderboard")

	return lb
}

// setChannel sets the channel ID for the leaderboard to publish the monthly leaderboard.
func (lb *Leaderboard) setChannel(channelID string) {
	lb.ChannelID = channelID
	writeLeaderboard(lb)
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
	log.WithFields(log.Fields{"guildID": lb.GuildID, "account": account, "rank": rank}).Debug("lifetime ranking")

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
	log.WithFields(log.Fields{"guildID": lb.GuildID, "account": account, "rank": rank}).Debug("lifetime ranking")

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
	log.WithFields(log.Fields{"guildID": lb.GuildID, "account": account, "rank": rank}).Debug("lifetime ranking")

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
			log.WithFields(log.Fields{"error": err, "leaderboard": accounts}).Error("unable to send montly leaderboard")
			return err
		}
	} else {
		log.WithField("guildID", lb.ChannelID).Warning("no leaderboard channel set for server")
	}
	for _, account := range accounts {
		log.WithFields(log.Fields{"guildID": lb.GuildID, "memberID": account.MemberID, "monthlyBalance": account.MonthlyBalance}).Info("sent monthly leaderboard")
	}
	log.WithFields(log.Fields{"guildID": lb.GuildID, "leaderboard": accounts}).Info("sent monthly leaderboard")
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
				log.WithFields(log.Fields{"guildID": lb.GuildID, "error": err}).Error("unable to send monthly leaderboard")
			}
			lb.LastSeason = disctime.NextMonth(lastSeason)
			writeLeaderboard(lb)
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
