package shop

import (
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PURCHASED = "purchased"
	PENDING   = "pending"
	APPROVED  = "approved"
	DENIED    = "denied"
)

// Purchasable is an interface that represents an item that can be purchased from the shop.
type Purchasable interface {
	GetID() primitive.ObjectID
	GetName() string
	GetPrice() int
	GetDescription() string
	Buy(string, string) (*Purchase, error)
}

// Purchase is a purchase made from the shop.
type Purchase struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	GuildID     string             `json:"guildID" bson:"guildID"`
	MemberID    string             `json:"memberID" bson:"memberID"`
	Purchasable Purchasable        `json:"purchasable" bson:"purchasable"`
	Status      string             `json:"status" bson:"status"`
	Date        time.Time          `json:"date" bson:"date"`
}

// GetAllPurchasableItems returns all items that may be purchased in the shop.
func GetAllPurchasableItems(guildID string) []Purchasable {
	log.Trace("--> shop.GetAllPurchasableItems")
	defer log.Trace("<-- shop.GetAllPurchasableItems")

	shopItems, err := readAllPurchasableItems(guildID)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "error": err}).Error("unable to read purchasable items from the database")
		return nil
	}
	return shopItems
}

// GetAllRoles returns all the purchases made by a member in the guild.
func GetAllPurchases(guildID string, memberID string) []*Purchase {
	log.Trace("--> shop.GetAllPurchases")
	defer log.Trace("<-- shop.GetAllPurchases")

	purchases, err := readAllPurchases(guildID, memberID)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "member": memberID, "error": err}).Error("unable to read purchases from the database")
		return nil
	}

	return purchases
}

// NewPurchase creates a new Purchase with the given guild ID, member ID, and Purchasable.
func NewPurchase(guildID, memberID string, purchasable Purchasable) (*Purchase, error) {
	purchase := &Purchase{
		GuildID:     guildID,
		MemberID:    memberID,
		Purchasable: purchasable,
		Status:      PENDING,
		Date:        time.Now(),
	}

	return purchase, nil
}
