package shop

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ROLE = "role"
)

// The shop for a guild. The shop contains all items available for purchase.
type Shop struct {
	GuildID string      // Guild (server) for the shop
	Items   []*ShopItem // All items available in the shop
}

type ShopItem struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID       string             `json:"guild_id" bson:"guild_id"`
	Name          string             `json:"name" bson:"name"`
	Description   string             `json:"description" bson:"description"`
	Type          string             `json:"type" bson:"type"`
	Price         int                `json:"price" bson:"price"`
	Duration      time.Duration      `json:"duration,omitempty" bson:"duration,omitempty"`
	AutoRenewable bool               `json:"auto_renewable,omitempty" bson:"auto_renewable,omitempty"`
}

// GetShop returns the shop for the guild.
func GetShop(guildID string) *Shop {
	log.Trace("--> shop.GetShop")
	defer log.Trace("<-- shop.GetShop")

	var err error

	shop := &Shop{
		GuildID: guildID,
	}

	shop.Items, err = readShopItems(guildID)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "error": err}).Error("unable to read shop items from the database")
		shop.Items = make([]*ShopItem, 0)
	}

	return shop
}

// GetShopItem returns the shop item with the given guild ID, name, and type. If the item does
// not exist, nil is returned.
func GetShopItem(guildID string, name string, itemType string) *ShopItem {
	log.Trace("--> shop.GetShopItem")
	defer log.Trace("<-- shop.GetShopItem")

	item, err := readShopItem(guildID, name, itemType)
	if err != nil || item == nil {
		return nil
	}

	return item
}

// NewShopItem creates a new ShopItem with the given guild ID, name, description, type, and price.
func NewShopItem(guildID string, name string, description string, itemType string, price int, duration time.Duration, autoRenewable bool) *ShopItem {
	// TODO: write to the DB, but verify it is a unique item (or simply update it if it already exists)
	//       the DB key should be guidID, name, and type.
	item := &ShopItem{
		GuildID:       guildID,
		Name:          name,
		Description:   description,
		Type:          itemType,
		Price:         price,
		Duration:      time.Duration(0),
		AutoRenewable: false,
	}

	err := writeShopItem(item)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "name": name, "type": itemType, "error": err}).Error("unable to write shop item to the database")
		return nil
	}

	log.WithFields(log.Fields{"guild": guildID, "name": name, "type": itemType}).Info("new shop item created")
	return item
}

// GetShopItems finds an item in the shop. If the item does not exist then nil is returned.
func (s *Shop) GetShopItem(name string, itemType string) *ShopItem {
	log.Trace("--> shop.GetShopItem")
	defer log.Trace("<-- shop.GetShopItem")

	for _, item := range s.Items {
		if item.Name == name && item.Type == itemType {
			return item
		}
	}

	return nil
}

// AddShopItem adds a new item to the shop. If the item already exists, an error is returned.
func (s *Shop) AddShopItem(name string, description string, itemType string, price int, duration time.Duration, renewable bool) (*ShopItem, error) {
	log.Trace("--> shop.AddShopItem")
	defer log.Trace("<-- shop.AddShopItem")

	item := s.GetShopItem(name, itemType)
	if item != nil {
		return nil, fmt.Errorf("item already exists")
	}

	item = NewShopItem(s.GuildID, name, description, itemType, price, duration, renewable)
	if item == nil {
		log.WithFields(log.Fields{"guild": s.GuildID, "name": name, "type": itemType}).Error("unable to write shop item to the database")
		return nil, fmt.Errorf("unable to add item")
	}
	s.Items = append(s.Items, item)

	log.WithFields(log.Fields{"guild": item.GuildID, "name": item.Name, "type": item.Type}).Info("shop item added")
	return item, nil
}

// RemoveShopItem removes an item from the shop. If the item does not exist, an error is returned.
func (s *Shop) RemoveShopItem(name string, itemType string) error {
	log.Trace("--> shop.RemoveShopItem")
	defer log.Trace("<-- shop.RemoveShopItem")

	item := s.GetShopItem(name, itemType)
	if item == nil {
		return fmt.Errorf("item not found")
	}
	err := deleteShopItem(item)
	if err != nil {
		return fmt.Errorf("unable to remove item")
	}

	log.WithFields(log.Fields{"guild": s.GuildID, "name": name, "type": itemType}).Info("shop item removed")
	return nil
}

// Purchase purchases the shop item for the given member. If the purchase is successful, a Purchase
// object is returned. If the purchase fails, an error is returned.
func (item *ShopItem) Purchase(memberID string, renew bool) (*Purchase, error) {
	log.Trace("--> shop.ShopItem.Purchase")
	defer log.Trace("<-- shop.ShopItem.Purchase")

	purchase, err := NewPurchase(item.GuildID, memberID, item, renew)
	if err != nil {
		log.WithFields(log.Fields{"guild": item.GuildID, "member": memberID, "item": item.Name, "error": err}).Error("unable to create purchase")
		return nil, fmt.Errorf("unable to purchase the item")
	}

	return purchase, nil
}

// UpdateShopItem updates the shop item with the given name and type. If the item does not exist, an error is returned.
func (item *ShopItem) UpdateShopItem(name string, description string, itemType string, price int, duration time.Duration, autoRenewable bool) error {
	log.Trace("--> shop.ShopItem.UpdateShopItem")
	defer log.Trace("<-- shop.ShopItem.UpdateShopItem")

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

// String returns a string representation of the Role.
func (item *ShopItem) String() string {
	return fmt.Sprintf("ShopItem{ID: %s, Role: %s Price: %d Description: %s}", item.ID.Hex(), item.Name, item.Price, item.Description)
}
