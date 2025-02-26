package payday

import (
	"fmt"
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
	log.Trace("--> payday.GetPayday")
	defer log.Trace("<-- payday.GetPayday")

	payday := readPayday(guildID)
	if payday == nil {
		payday = newPayday(guildID)
	}

	return payday
}

// GetAccount returns an account in the guild (server). If one doesn't exist, then nil is returned.
func (payday *Payday) GetAccount(memberID string) *Account {
	log.Trace("--> payday.Payday.getAccount")
	defer log.Trace("<-- payday.Payday.getAccount")

	return getAccount(payday, memberID)
}

// SetPaydayAmount sets the amount of credits a player deposits into their account on a given payday.
func (payday *Payday) SetPaydayAmount(amount int) {
	log.Trace("--> payday.SetPaydayAmount")
	defer log.Trace("<-- payday.SetPaydayAmount")

	payday.Amount = amount

	writePayday(payday)
}

// SetPaydayFrequency sets the frequency of paydays at which a player can deposit credits into their account.
func (payday *Payday) SetPaydayFrequency(frequency time.Duration) {
	log.Trace("--> payday.SetPaydayFrequency")
	defer log.Trace("<-- payday.SetPaydayFrequency")

	payday.PaydayFrequency = frequency

	writePayday(payday)
}

// newPayday creates new payday information for a server/guild
func newPayday(guildID string) *Payday {
	log.Trace("--> payday.newPayday")
	defer log.Trace("<-- payday.newPayday")

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
