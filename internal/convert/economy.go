package convert

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/database/mongo"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/rbrabson/goblin/leaderboard"
)

func ConvertEconomy(fileName string) {
	bytes := readFile(fileName)
	fileContents := asArray(bytes)
	for _, fileContent := range fileContents {
		guildID := asString(fileContent["_id"])
		if guildID == GUILD_ID {
			convertBankAccountModel(guildID, fileContent)
		}
	}
}

func convertBankAccountModel(guildID string, model map[string]interface{}) {
	db := mongo.NewDatabase()

	accounts := convertBankAccounts(guildID, model)
	for _, account := range accounts {
		filter := bson.M{"guild_id": account.GuildID, "member_id": account.MemberID}
		err := db.UpdateOrInsert(bank.ACCOUNT_COLLECTION, filter, account)
		if err != nil {
			fmt.Printf("error inserting account: %v\n", err)
		} else {
			fmt.Printf("inserted account: %v\n", account)
		}
	}
}

func convertBank(guildID string, model map[string]interface{}) {
	name := asString(model["bank_name"])
	currency := asString(model["currency"])
	defaultBalance := asInteger(model["default_balance"])

	type bank struct {
		GuildID        string `json:"guild_id"`
		Name           string `json:"name"`
		Currency       string `json:"currency"`
		DefaultBalance int    `json:"default_balance"`
	}
	b := &bank{
		GuildID:        guildID,
		Name:           name,
		Currency:       currency,
		DefaultBalance: defaultBalance,
	}
	bankJson, _ := json.Marshal(b)
	fmt.Println(bankJson)

	lastSeason := time.Time{}
	channelID := asString(model["channel_id"])
	fmt.Printf("guildID=%s, bankName=%s, currency=%s, defaultBalance=%d, lastSeason=%v, channelID=%s\n", guildID, name, currency, defaultBalance, lastSeason, channelID)
}

func convertBankAccounts(guildID string, model map[string]interface{}) []*bank.Account {
	accounts := make([]*bank.Account, 0)
	bankAccounts := asMap(model["accounts"])
	for _, bankAccount := range bankAccounts {
		accounts = append(accounts, convertBankAccount(guildID, asMap(bankAccount)))
	}
	return accounts
}

func convertBankAccount(guildID string, model map[string]interface{}) *bank.Account {
	memberID := asString(model["_id"])
	monthlyBalance := asInteger(model["monthly_balance"])
	currentBalance := asInteger(model["current_balance"])
	lifetimeBalance := asInteger(model["lifetime_balance"])
	createdAt := asTime(model["created_at"])

	account := &bank.Account{
		GuildID:         guildID,
		MemberID:        memberID,
		MonthlyBalance:  monthlyBalance,
		CurrentBalance:  currentBalance,
		LifetimeBalance: lifetimeBalance,
		CreatedAt:       createdAt,
	}
	return account
}

func convertLeaderboard(guildID string, model map[string]interface{}) {
	channelID := asString(model["channel_id"])
	lastSeason := asTime(model["last_season"])

	lb := leaderboard.Leaderboard{
		GuildID:    guildID,
		ChannelID:  channelID,
		LastSeason: lastSeason,
	}
	fmt.Println(lb.String())
}
