package shop

import "time"

const (
	ROLE_TYPE = "role"
)

// Role is a custom role that a member can purchase in the shop.
type Role ShopItem

// NewRole returns a new role that can be purchased in the shop
func NewRole(guildID string, name string, price int, description string, duration time.Duration, renewable bool) *ShopItem {
	return NewShopItem(guildID, name, description, ROLE_TYPE, price, duration, renewable)
}

func (r *Role) Buy(guildID, memberID string, autoRenew bool) (*Purchase, error) {
	// TODO: verify the member has sufficient funds, and withdraw them from
	//       their bank account.
	// TODO: verify the user hasn't already bought the same exact item
	//       in the shop. For instance, don't buy a role if you already
	//       have it.
	// TODO: assign the role to the user in the guild.
	item := ShopItem(*r)
	return NewPurchase(guildID, memberID, &item, autoRenew)
}
