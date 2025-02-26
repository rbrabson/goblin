package convert

import (
	"fmt"
	"time"

	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/leaderboard"
)

func ConvertEconomy(fileName string) {
	bytes := readFile(fileName)
	fileContents := asArray(bytes)
	for _, fileContent := range fileContents {
		guildID := asString(fileContent["_id"])
		if guildID == GUILD_ID {
			convertEconomyModel(guildID, fileContent)
		}
	}
}

func convertEconomyModel(guildID string, model map[string]interface{}) {
	convertBank(guildID, model)
	convertLeaderboard(guildID, model)
	// convertBankAccounts(guildID, model)
}

func convertBank(guildID string, model map[string]interface{}) {
	name := asString(model["bank_name"])
	currency := asString(model["currency"])
	defaultBalance := asInteger(model["default_balance"])

	b := &bank.Bank{
		GuildID:        guildID,
		Name:           name,
		Currency:       currency,
		DefaultBalance: defaultBalance,
	}
	fmt.Println(b.String())

	lastSeason := time.Time{}
	channelID := asString(model["channel_id"])
	fmt.Printf("guildID=%s, bankName=%s, currency=%s, defaultBalance=%d, lastSeason=%v, channelID=%s\n", guildID, name, currency, defaultBalance, lastSeason, channelID)
}

func convertBankAccounts(guildID string, model map[string]interface{}) {
	bankAccounts := asMap(model["accounts"])
	for _, bankAccount := range bankAccounts {
		convertBankAccount(guildID, asMap(bankAccount))
	}
}

func convertBankAccount(guildID string, model map[string]interface{}) {
	memberID := asString(model["_id"])
	monthlyBalance := asInteger(model["monthly_balance"])
	currentBalance := asInteger(model["current_balance"])
	lifetimeBalance := asInteger(model["lifetime_balance"])
	createdAt := asTime(model["created_at"])

	fmt.Printf("guildID=%s, memberID=%s, monthlyBalance=%d, currentBalance=%d, lifetimeBalance=%d, createdAt=%v\n", guildID, memberID, monthlyBalance, currentBalance, lifetimeBalance, createdAt)
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
