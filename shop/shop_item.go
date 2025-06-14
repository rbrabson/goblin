package shop

import (
	"fmt"
	"log/slog"

	"github.com/rbrabson/goblin/bank"
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
	MaxPurhases   int                `json:"max_purchases,omitempty" bson:"max_purchases,omitempty"`
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
func newShopItem(guildID string, name string, description string, itemType string, price int, duration string, autoRenewable bool, maxPurchases int) *ShopItem {
	item := &ShopItem{
		GuildID:       guildID,
		Name:          name,
		Description:   description,
		Type:          itemType,
		Price:         price,
		Duration:      duration,
		AutoRenewable: autoRenewable,
		MaxPurhases:   maxPurchases,
	}

	err := writeShopItem(item)
	if err != nil {
		slog.Error("unable to write shop item to the database",
			slog.String("guild", guildID),
			slog.String("name", name),
			slog.String("type", itemType),
			slog.Any("error", err),
		)
		return nil
	}

	slog.Info("new shop item created",
		slog.String("guild", guildID),
		slog.String("name", name),
		slog.String("type", itemType),
	)

	return item
}

// update updates the shop item with the given name and type. If the item does not exist, an error is returned.
func (item *ShopItem) update(name string, description string, itemType string, price int, duration string, autoRenewable bool) error {
	if item.Name == name && item.Description == description && item.Type == itemType && item.Price == price && duration == item.Duration && autoRenewable == item.AutoRenewable {
		slog.Warn("no change to the shop item",
			slog.String("guild", item.GuildID),
			slog.String("name", item.Name),
			slog.String("type", item.Type),
		)
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
		slog.Error("unable to update shop item to the database",
			slog.String("guild", item.GuildID),
			slog.String("name", item.Name),
			slog.String("type", item.Type),
			slog.Any("error", err),
		)
		return fmt.Errorf("unable to add item")
	}
	slog.Info("shop item updated",
		slog.String("guild", item.GuildID),
		slog.String("name", item.Name),
		slog.String("type", item.Type),
	)
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
		slog.Error("unable to write shop item to the database",
			slog.String("guild", item.GuildID),
			slog.String("name", item.Name),
			slog.String("type", item.Type),
			slog.Any("error", err),
		)
		return fmt.Errorf("unable to add %s to shop", item.Type)
	}

	s.Items = append(s.Items, item)
	slog.Info("shop item added to shop",
		slog.String("guild", item.GuildID),
		slog.String("name", item.Name),
		slog.String("type", item.Type),
	)
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
		slog.Error("unable to remove shop item from the database",
			slog.String("guild", item.GuildID),
			slog.String("name", item.Name),
			slog.String("type", item.Type),
			slog.Any("error", err),
		)
		return fmt.Errorf("unable to remove %s from shop", item.Type)
	}

	// Remove the item from the shop's items slice
	for i, it := range s.Items {
		if it.ID == item.ID {
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
			break
		}
	}

	slog.Info("shop item removed from shop",
		slog.String("guild", item.GuildID),
		slog.String("name", item.Name),
		slog.String("type", item.Type),
	)
	return nil
}

// purchase purchases the shop item for the given member. If the purchase is successful, a purchase
// object is returned. If the purchase fails, an error is returned.
func (item *ShopItem) purchase(memberID string, status string, renew bool) (*Purchase, error) {
	purchase, err := PurchaseItem(item.GuildID, memberID, item, status, renew)
	if err != nil {
		slog.Error("unable to create purchase",
			slog.String("guild", item.GuildID),
			slog.String("member", memberID),
			slog.String("item", item.Name),
			slog.Any("error", err),
		)
		return nil, err
	}

	return purchase, nil
}

// createChecks performs checks to see if a role can be added to the shop.
func createChecks(guildID string, itemName string, itemType string) error {
	shopItem := getShopItem(guildID, itemName, itemType)
	if shopItem != nil {
		slog.Error("item already exists in the shop",
			slog.String("guild", guildID),
			slog.String("name", itemName),
			slog.String("type", itemType),
		)
		return fmt.Errorf("%s `%s` already exists in the shop", itemType, itemName)
	}

	return nil
}

// purchaseChecks performs checks to see if a member can purchase the shop item.
func purchaseChecks(guildID string, memberID string, itemType string, itemName string) error {
	purchase, _ := readPurchase(guildID, memberID, itemName, itemType)
	if purchase != nil && !purchase.IsExpired {
		slog.Debug("item already purchased",
			slog.String("guild", guildID),
			slog.String("member", memberID),
			slog.String("name", itemName),
			slog.String("type", itemType),
		)
		return fmt.Errorf("you have already purchased %s `%s`", itemType, itemName)
	}

	// Make sure the member has sufficient funds to purchase the item
	item := getShopItem(guildID, itemName, itemType)
	bankAccount := bank.GetAccount(guildID, memberID)
	if bankAccount.CurrentBalance < item.Price {
		slog.Debug("insufficient funds to purchase item",
			slog.String("guild", guildID),
			slog.String("name", itemName),
			slog.String("type", itemType),
			slog.String("member", memberID),
		)
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
