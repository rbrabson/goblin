package shop

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Shop struct{}

type ShopItem struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID     string             `json:"guildID" bson:"guildID"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	Type        string             `json:"type" bson:"type"`
	Price       int                `json:"price" bson:"price"`
}

// GetShopItem returns the shop item with the given guild ID, name, and type.
func GetShopItem(guildID string, name string, itemType string) *ShopItem {
	// TODO: read from the DB, using "guildID, name, type as the keys"
	return nil
}

// GetAllShopItems returns all the shop items for the given guild ID.
func GetAllShopItems(guildID string) []*ShopItem {
	// TODO: read from the DB, using "guildID" as the key
	return nil
}

// NewShopItem creates a new ShopItem with the given guild ID, name, description, type, and price.
func NewShopItem(guildID string, name string, description string, itemType string, price int) *ShopItem {
	// TODO: write to the DB, but verify it is a unique item (or simply update it if it already exists)
	//       the DB key should be guidID, name, and type.
	return &ShopItem{
		GuildID:     guildID,
		Name:        name,
		Description: description,
		Type:        itemType,
		Price:       price,
	}
}

// String returns a string representation of the Role.
func (item *ShopItem) String() string {
	return fmt.Sprintf("ID: %s, Role: %s Price: %d Description: %s", item.ID.Hex(), item.Name, item.Price, item.Description)
}
