package shop

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	APPROVED  = "approved"
	DENIED    = "denied"
	PENDING   = "pending"
	PURCHASED = "purchased"
)

// Purchase is a purchase made from the shop.
type Purchase struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	GuildID     string             `json:"guildID" bson:"guildID"`
	MemberID    string             `json:"memberID" bson:"memberID"`
	Item        *ShopItem          `json:"item" bson:"item"`
	Status      string             `json:"status" bson:"status"`
	PurchasedOn time.Time          `json:"purchased_on" bson:"purchased_on"`
	ExpiresOn   time.Time          `json:"expires_on" bson:"expires_on"`
	AutoRenew   bool               `json:"autoRenew" bson:"autoRenew"`
}

// GetAllPurchasableItems returns all items that may be purchased in the shop.
func GetAllPurchasableItems(guildID string) []*ShopItem {
	log.Trace("--> shop.GetAllPurchasableItems")
	defer log.Trace("<-- shop.GetAllPurchasableItems")

	shopItems, err := readShopItems(guildID)
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

	purchases, err := readPurchases(guildID, memberID)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "member": memberID, "error": err}).Error("unable to read purchases from the database")
		return nil
	}

	return purchases
}

// NewPurchase creates a new Purchase with the given guild ID, member ID, and a purchasable
// shop item.
func NewPurchase(guildID, memberID string, item *ShopItem, renew bool) (*Purchase, error) {
	purchase := &Purchase{
		GuildID:     guildID,
		MemberID:    memberID,
		Item:        item,
		Status:      PENDING,
		PurchasedOn: time.Now(),
		ExpiresOn:   time.Now().Add(item.Duration),
		AutoRenew:   renew,
	}
	err := writePurchase(purchase)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "member": memberID, "item": item.Name, "error": err}).Error("unable to write purchase to the database")
		return nil, fmt.Errorf("unable to write purchase to the database: %w", err)
	}
	log.WithFields(log.Fields{"guild": guildID, "member": memberID, "item": item.Name}).Info("creating new purchase")

	return purchase, nil
}
