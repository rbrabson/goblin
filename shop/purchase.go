package shop

import (
	"fmt"
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

// Purchase is a purchase made from the shop.
type Purchase struct {
	ID       primitive.ObjectID `json:"_id" bson:"_id"`
	GuildID  string             `json:"guildID" bson:"guildID"`
	MemberID string             `json:"memberID" bson:"memberID"`
	Item     *ShopItem          `json:"item" bson:"item"`
	Status   string             `json:"status" bson:"status"`
	Date     time.Time          `json:"date" bson:"date"`
}

// GetAllPurchasableItems returns all items that may be purchased in the shop.
func GetAllPurchasableItems(guildID string) []*ShopItem {
	log.Trace("--> shop.GetAllPurchasableItems")
	defer log.Trace("<-- shop.GetAllPurchasableItems")

	shopItems, err := readAllShopItems(guildID)
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
func NewPurchase(guildID, memberID string, item *ShopItem) (*Purchase, error) {

	purchase, _ := readPurchase(guildID, memberID, item.Name)
	if purchase != nil {
		log.WithFields(log.Fields{"guild": guildID, "member": memberID, "item": item.Name}).Error("already purchases the item")
		return nil, fmt.Errorf("purchase already exists for item %s", item.Name)
	}

	purchase = &Purchase{
		GuildID:  guildID,
		MemberID: memberID,
		Item:     item,
		Status:   PENDING,
		Date:     time.Now(),
	}
	err := writePurchase(purchase)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "member": memberID, "item": item.Name, "error": err}).Error("unable to write purchase to the database")
		return nil, fmt.Errorf("unable to write purchase to the database: %w", err)
	}
	log.WithFields(log.Fields{"guild": guildID, "member": memberID, "item": item.Name}).Info("creating new purchase")

	return purchase, nil
}
