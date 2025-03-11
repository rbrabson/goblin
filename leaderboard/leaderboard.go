package leaderboard

import (
	"time"

	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/internal/discmsg"
	"github.com/rbrabson/goblin/internal/disctime"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/text/language"
)

// A Leaderboard is used to send a monthly leaderboard to the Discord server for each guild.
type Leaderboard struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID    string             `json:"guild_id" bson:"guild_id"`
	ChannelID  string             `json:"channel_id" bson:"channel_id"`
	LastSeason time.Time          `json:"last_season" bson:"last_season"`
}

func newLeaderboard(guildID string) *Leaderboard {
	log.Trace("--> leaderboard.newLeaderboard")
	defer log.Trace("<-- leaderboard.newLeaderboard")

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
	log.Trace("--> leaderboard.getLeaderboards")
	defer log.Trace("<-- leaderboard.getLeaderboards")

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
	log.Trace("--> leaderboard.getLeaderboard")
	defer log.Trace("<-- leaderboard.getLeaderboard")

	lb := readLeaderboard(guildID)
	if lb == nil {
		lb = newLeaderboard(guildID)
	}
	log.WithFields(log.Fields{"guildID": guildID, "leaderboard": lb}).Trace("leaderboard")

	return lb
}

func (lb *Leaderboard) setChannel(channelID string) {
	log.Trace("--> leaderboard.setChannel")
	defer log.Trace("<-- leaderboard.setChannel")

	lb.ChannelID = channelID
	writeLeaderboard(lb)
}

// GetCurrentRanking returns the global rankings based on the current balance.
func (lb *Leaderboard) getCurrentLeaderboard() []*bank.Account {
	log.Trace("--> leaderboard.getCurrentLeaderboard")
	defer log.Trace("<-- leaderboard.getCurrentLeaderboard")

	filter := bson.D{{Key: "guild_id", Value: lb.GuildID}}
	sort := bson.D{{Key: "current_balance", Value: -1}, {Key: "_id", Value: 1}}
	limit := int64(10)

	accounts := bank.GetAccounts(lb.GuildID, filter, sort, limit)

	return accounts
}

// getMonthlyLeaderboard returns the global rankings based on the monthly balance.
func (lb *Leaderboard) getMonthlyLeaderboard() []*bank.Account {
	log.Trace("--> leaderboard.getMonthlyLeaderboard")
	defer log.Trace("<-- leaderboard.getMonthlyLeaderboard")

	filter := bson.D{{Key: "guild_id", Value: lb.GuildID}}
	sort := bson.D{{Key: "monthly_balance", Value: -1}, {Key: "_id", Value: 1}}
	limit := int64(10)

	accounts := bank.GetAccounts(lb.GuildID, filter, sort, limit)

	return accounts
}

// getLifetimeLeaderboard returns the global rankings based on the monthly balance.
func (lb *Leaderboard) getLifetimeLeaderboard() []*bank.Account {
	log.Trace("--> leaderboard.getLifetimeLeaderboard")
	defer log.Trace("<-- leaderboard.getLifetimeLeaderboard")

	filter := bson.D{{Key: "guild_id", Value: lb.GuildID}}
	sort := bson.D{{Key: "lifetime_balance", Value: -1}, {Key: "_id", Value: 1}}
	limit := int64(10)

	accounts := bank.GetAccounts(lb.GuildID, filter, sort, limit)

	return accounts
}

// getMonthlyRanking returns the monthly global ranking on the server for a given player.
func getCurrentRanking(lb *Leaderboard, account *bank.Account) int {
	log.Trace("--> leaderboard.getCurrentyRanking")
	defer log.Trace("<-- leaderboard.getCurrentRanking")

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
	log.Trace("--> leaderboard.getMonthlyRanking")
	defer log.Trace("<-- leaderboard.getMonthlyRanking")

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
	log.Trace("--> leaderboard.getLifetimeRanking")
	defer log.Trace("<-- leaderboard.getLifetimeRanking")

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
func sendhMonthlyLeaderboard(lb *Leaderboard) error {
	log.Trace("--> bank.sendMonthlyLeaderboard")
	defer log.Trace("<-- bank.sendMonthlyLeaderboard")

	// Get the top 10 accounts for this month
	sortedAccounts := lb.getMonthlyLeaderboard()
	leaderboardSize := min(10, len(sortedAccounts))
	sortedAccounts = sortedAccounts[:leaderboardSize]

	firstOfMonth := disctime.PreviousMonth(time.Now())
	year, month, _ := firstOfMonth.Date()
	if lb.ChannelID != "" {
		p := discmsg.GetPrinter(language.AmericanEnglish)
		embeds := formatAccounts(p, fmt.Sprintf("%s %d Top 10", month, year), sortedAccounts)
		_, err := bot.Session.ChannelMessageSendComplex(lb.ChannelID, &discordgo.MessageSend{
			Embeds: embeds,
		})
		if err != nil {
			log.WithFields(log.Fields{"error": err, "leaderboard": sortedAccounts}).Error("unable to send montly leaderboard")
			return err
		}
	} else {
		log.WithField("guildID", lb.ChannelID).Warning("no leaderboard channel set for server")
	}
	return nil
}

// publishMonthlyLeaderboard sends the monthly leaderboard to each guild.
func sendMonthlyLeaderboard() {
	log.Trace("--> bank.sendMonthlyLeaderboard")
	defer log.Trace("<-- bank.sendMonthlyLeaderboard")

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
		lastSeason = disctime.CurrentMonth(time.Now())

		leaderboards := getLeaderboards()
		for _, lb := range leaderboards {
			err := sendhMonthlyLeaderboard(lb)
			if err != nil {
				log.WithFields(log.Fields{"guildID": lb.GuildID, "error": err}).Error("unable to send monthly leaderboard")
			}
			lb.LastSeason = lastSeason
			writeLeaderboard(lb)
		}

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
