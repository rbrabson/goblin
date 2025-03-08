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
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID     string             `json:"guild_id" bson:"guild_id"`
	MemberID    string             `json:"member_id" bson:"member_id"`
	Item        *ShopItem          `json:"item" bson:"item,inline"`
	Status      string             `json:"status" bson:"status"`
	PurchasedOn time.Time          `json:"purchased_on" bson:"purchased_on"`
	ExpiresOn   time.Time          `json:"expires_on,omitempty" bson:"expires_on,omitempty"`
	AutoRenew   bool               `json:"autoRenew,omitempty" bson:"autoRenew,omitempty"`
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
		Status:      PURCHASED,
		PurchasedOn: time.Now(),
	}
	if item.AutoRenewable {
		purchase.AutoRenew = renew
	}
	if item.Duration != 0 {
		purchase.ExpiresOn = time.Now().Add(item.Duration)
	}
	err := writePurchase(purchase)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "member": memberID, "item": item.Name, "error": err}).Error("unable to write purchase to the database")
		return nil, fmt.Errorf("unable to write purchase to the database: %w", err)
	}
	log.WithFields(log.Fields{"guild": guildID, "member": memberID, "item": item.Name}).Info("creating new purchase")

	return purchase, nil
}

// Update updates the purchase with the given autoRenew value.
func (p *Purchase) Update(autoRenew bool) error {
	log.Trace("--> shop.Purchase.Update")
	defer log.Trace("<-- shop.Purchase.Update")

	if p.AutoRenew == autoRenew {
		log.WithFields(log.Fields{"guild": p.GuildID, "member": p.MemberID, "item": p.Item.Name}).Info("purchase already has the same autoRenew value")
		return fmt.Errorf("purchase already has the same autoRenew value")
	}

	p.AutoRenew = autoRenew
	err := writePurchase(p)
	if err != nil {
		log.WithFields(log.Fields{"guild": p.GuildID, "member": p.MemberID, "item": p.Item.Name, "error": err}).Error("unable to update purchase in the database")
		return fmt.Errorf("unable to update purchase in the database: %w", err)
	}
	log.WithFields(log.Fields{"guild": p.GuildID, "member": p.MemberID, "item": p.Item.Name}).Info("updating purchase")

	return nil
}

// String returns a string representation of the purchase.
func (p *Purchase) String() string {
	return fmt.Sprintf("Purchase{GuildID: %s, MemberID: %s, Item: %v, Status: %s, PurchasedOn: %s, ExpiresOn: %s, AutoRenew: %t}",
		p.Item.GuildID,
		p.MemberID,
		p.Item,
		p.Status,
		p.PurchasedOn.Format(time.RFC3339),
		p.ExpiresOn.Format(time.RFC3339),
		p.AutoRenew,
	)
}
