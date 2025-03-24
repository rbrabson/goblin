package shop

import (
	"fmt"

	"github.com/rbrabson/goblin/bank"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ShopItem represents an item in the shop, which represents any purchasable item.
type ShopItem struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID       string             `json:"guild_id" bson:"guild_id"`
	Name          string             `json:"name" bson:"name"`
	Description   string             `json:"description" bson:"description"`
	Type          string             `json:"type" bson:"type"`
	Price         int                `json:"price" bson:"price"`
	Duration      string             `json:"duration,omitempty" bson:"duration,omitempty"`
	AutoRenewable bool               `json:"auto_renewable,omitempty" bson:"auto_renewable,omitempty"`
}

// getShopItem returns the shop item with the given guild ID, name, and type. If the item does
// not exist, nil is returned.
func getShopItem(guildID string, name string, itemType string) *ShopItem {
	item, err := readShopItem(guildID, name, itemType)
	if err != nil || item == nil {
		return nil
	}

	return item
}

// newShopItem creates a new ShopItem with the given guild ID, name, description, type, and price.
func newShopItem(guildID string, name string, description string, itemType string, price int, duration string, autoRenewable bool) *ShopItem {
	item := &ShopItem{
		GuildID:       guildID,
		Name:          name,
		Description:   description,
		Type:          itemType,
		Price:         price,
		Duration:      duration,
		AutoRenewable: autoRenewable,
	}

	err := writeShopItem(item)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "name": name, "type": itemType, "error": err}).Error("unable to write shop item to the database")
		return nil
	}

	log.WithFields(log.Fields{"guild": guildID, "name": name, "type": itemType}).Info("new shop item created")
	return item
}

// update updates the shop item with the given name and type. If the item does not exist, an error is returned.
func (item *ShopItem) update(name string, description string, itemType string, price int, duration string, autoRenewable bool) error {
	if item.Name == name && item.Description == description && item.Type == itemType && item.Price == price && duration == item.Duration && autoRenewable == item.AutoRenewable {
		log.WithFields(log.Fields{"guild": item.GuildID, "name": item.Name, "type": item.Type}).Warn("no change to the shop item")
		return fmt.Errorf("no change to the shop item")
	}

	item.Name = name
	item.Description = description
	item.Type = itemType
	item.Price = price
	item.Duration = duration
	item.AutoRenewable = autoRenewable

	err := writeShopItem(item)
	if err != nil {
		log.WithFields(log.Fields{"guild": item.GuildID, "name": item.Name, "type": item.Type, "error": err}).Error("unable to update shop item to the database")
		return fmt.Errorf("unable to add item")
	}
	log.WithFields(log.Fields{"guild": item.GuildID, "name": item.Name, "type": item.Type}).Info("shop item updated")
	return nil
}

// addToShop adds the shop item to the given shop. If the item already exists in the shop, an error is returned.
func (item *ShopItem) addToShop(s *Shop) error {
	existingItem := s.GetShopItem(item.Name, item.Type)
	if existingItem != nil {
		return fmt.Errorf("%s already exists in the shop", item.Type)
	}

	err := writeShopItem(item)
	if err != nil {
		log.WithFields(log.Fields{"guild": item.GuildID, "name": item.Name, "type": item.Type, "error": err}).Error("unable to write shop item to the database")
		return fmt.Errorf("unable to add %s to shop", item.Type)
	}

	s.Items = append(s.Items, item)
	log.WithFields(log.Fields{"guild": item.GuildID, "name": item.Name, "type": item.Type}).Info("shop item added to shop")
	return nil
}

// removeFromShop removes the shop item from the given shop. If the item does not exist in the shop, an error is returned.
func (item *ShopItem) removeFromShop(s *Shop) error {
	// Check if the item exists in the shop
	existingItem := s.GetShopItem(item.Name, item.Type)
	if existingItem == nil {
		return fmt.Errorf("%s does not exist in the shop", item.Type)
	}

	// Remove the item from the database
	err := deleteShopItem(item)
	if err != nil {
		log.WithFields(log.Fields{"guild": item.GuildID, "name": item.Name, "type": item.Type, "error": err}).Error("unable to remove shop item from the database")
		return fmt.Errorf("unable to remove %s from shop", item.Type)
	}

	// Remove the item from the shop's items slice
	for i, it := range s.Items {
		if it.ID == item.ID {
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
			break
		}
	}

	log.WithFields(log.Fields{"guild": item.GuildID, "name": item.Name, "type": item.Type}).Info("shop item removed from shop")
	return nil
}

// purchase purchases the shop item for the given member. If the purchase is successful, a purchase
// object is returned. If the purchase fails, an error is returned.
func (item *ShopItem) purchase(memberID string, renew bool) (*Purchase, error) {
	purchase, err := PurchaseItem(item.GuildID, memberID, item, renew)
	if err != nil {
		log.WithFields(log.Fields{"guild": item.GuildID, "member": memberID, "item": item.Name, "error": err}).Error("unable to create purchase")
		return nil, err
	}

	return purchase, nil
}

// purchaseChecks performs checks to see if a member can purchase the shop item.
func purchaseChecks(guildID string, memberID string, itemType string, itemName string) error {
	purchase, _ := readPurchase(guildID, memberID, itemName, itemType)
	if purchase != nil && !purchase.IsExpired {
		log.WithFields(log.Fields{"guildID": guildID, "memberID": memberID, "itemName": itemName, "itemType": itemType}).Debug("item already purchased")
		return fmt.Errorf("you have already purchased %s `%s`", itemType, itemName)
	}

	// Make sure the member has sufficient funds to purchase the item
	item := getShopItem(guildID, itemName, itemType)
	bankAccount := bank.GetAccount(guildID, memberID)
	if bankAccount.CurrentBalance < item.Price {
		log.WithFields(log.Fields{"guildID": guildID, "name": itemName, "type": itemType, "memberID": memberID}).Debug("insufficient funds to purchase item")
		return fmt.Errorf("you do not have enough credits to purchase the `%s` %s", itemName, itemType)
	}
	return nil
}

// deleteShopItem deletes the shop item from the database. If the item does not exist, an error is returned.

// String returns a string representation of the Role.
func (item *ShopItem) String() string {
	return fmt.Sprintf("ShopItem{ID: %s, Guild: %s, Type: %s, Name: %s, Price: %d Description: %s, Duration: %s, AutoRenewable: %t}",
		item.ID.Hex(),
		item.GuildID,
		item.Type,
		item.Name,
		item.Price,
		item.Description,
		item.Duration,
		item.AutoRenewable,
	)
}
