package shop

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Role is a custom role that a member can purchase in the shop.
type Role struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Price       int                `json:"price" bson:"price"`
	Description string             `json:"description" bson:"description"`
}

func (r *Role) Buy(guildID, memberID string) (*Purchase, error) {
	purchasable := Purchasable(r)
	return NewPurchase(guildID, memberID, purchasable)
}

// GetID returns the ID of the Role.
func (r *Role) GetID() primitive.ObjectID {
	return r.ID
}

// GetName returns the name of the Role.
func (r *Role) GetName() string {
	return r.Name
}

// GetPrice returns the price of the Role.
func (r *Role) GetPrice() int {
	return r.Price
}

// GetDescription returns the description of the Role.
func (r *Role) GetDescription() string {
	return r.Description
}

// String returns a string representation of the Role.
func (r *Role) String() string {
	return fmt.Sprintf("ID: %s, Role: %s Price: %d Description: %s", r.ID.Hex(), r.Name, r.Price, r.Description)
}
