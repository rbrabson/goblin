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

func GetShopItem(guildID string, name string, itemType string) *ShopItem {
	// TODO: read from the DB, using "guildID, name, type as the keys"
	return nil
}

// String returns a string representation of the Role.
func (item *ShopItem) String() string {
	return fmt.Sprintf("ID: %s, Role: %s Price: %d Description: %s", item.ID.Hex(), item.Name, item.Price, item.Description)
}
