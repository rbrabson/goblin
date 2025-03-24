package payday

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	DEFAULT_PAYDAY_AMOUNT    = 5000
	DEFAULT_PAYDAY_FREQUENCY = time.Duration(23 * time.Hour)
)

// Payday is the daily payment for members of a guild (server).
type Payday struct {
	ID              primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID         string             `json:"guild_id" bson:"guild_id"`
	Amount          int                `json:"payday_amount" bson:"payday_amount"`
	PaydayFrequency time.Duration      `json:"payday_frequency" bson:"payday_frequency"`
}

// GetPayday returns the payday information for a server, creating a new one if necessary.
func GetPayday(guildID string) *Payday {
	payday := readPayday(guildID)
	if payday == nil {
		payday = readPaydayFromFile(guildID)
	}

	return payday
}

// GetAccount returns an account in the guild (server). If one doesn't exist, then nil is returned.
func (payday *Payday) GetAccount(memberID string) *Account {
	account := readAccount(payday, memberID)

	if account == nil {
		account = newAccount(payday, memberID)
	}

	return account
}

// SetPaydayAmount sets the amount of credits a player deposits into their account on a given payday.
func (payday *Payday) SetPaydayAmount(amount int) {
	payday.Amount = amount

	writePayday(payday)
}

// SetPaydayFrequency sets the frequency of paydays at which a player can deposit credits into their account.
func (payday *Payday) SetPaydayFrequency(frequency time.Duration) {
	payday.PaydayFrequency = frequency

	writePayday(payday)
}

// readPaydayFromFile creates new payday information for a server/guild.
// If the default payday configuration file cannot be read or dedcoded, then a
// default payday configuration is created.
func readPaydayFromFile(guildID string) *Payday {
	configTheme := os.Getenv("DISCORD_DEFAULT_THEME")
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	configFileName := filepath.Join(configDir, "payday", "config", configTheme+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		log.WithField("file", configFileName).Error("failed to read default payday config")
		return getDefaultPayday(guildID)
	}

	payday := &Payday{}
	err = json.Unmarshal(bytes, payday)
	if err != nil {
		log.WithField("file", configFileName).Error("failed to unmarshal default payday config")
		return getDefaultPayday(guildID)
	}
	payday.GuildID = guildID

	writePayday(payday)
	log.WithField("guild", payday.GuildID).Info("create new payday config")

	return payday
}

// getDefaultPayday creates new payday information for a server/guild
func getDefaultPayday(guildID string) *Payday {
	payday := &Payday{
		GuildID:         guildID,
		Amount:          DEFAULT_PAYDAY_AMOUNT,
		PaydayFrequency: DEFAULT_PAYDAY_FREQUENCY,
	}
	writePayday(payday)
	log.WithFields(log.Fields{"payday": payday}).Debug("created new payday")

	return payday
}

// String returns a string representation of the Payday.
func (payday *Payday) String() string {
	return fmt.Sprintf("Payday{ID=%s, GuildID=%s, Amount=%d, PaydayFrequency=%s}",
		payday.ID.Hex(),
		payday.GuildID,
		payday.Amount,
		payday.PaydayFrequency,
	)
}
