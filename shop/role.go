package shop

import log "github.com/sirupsen/logrus"

const (
	ROLE = "role"
)

// Role represents a role item in the shop.
type Role ShopItem

// GetRole retrieves a role from the shop by its name for a specific guild.
func GetRole(guildID string, name string) *Role {
	log.Trace("--> shop.GetRole")
	defer log.Trace("<-- shop.GetRole")

	item := GetShopItem(guildID, name, ROLE)
	role := Role(*item)
	return &role
}

// NewRole creates a new role for the shop.
func NewRole(guildID string, name string, description string, price int, duration string, autoRenewable bool) *Role {
	log.Trace("--> shop.NewRole")
	defer log.Trace("<-- shop.NewRole")

	item := NewShopItem(guildID, name, description, ROLE, price, duration, autoRenewable)
	role := (*Role)(item)
	return role
}

// Purchase allows a member to purchase the role from the shop.
func (r *Role) Purchase(memberID string, renew bool) (*Purchase, error) {
	log.Trace("--> shop.Role.Purchase")
	defer log.Trace("<-- shop.Role.Purchase")

	item := ShopItem(*r)
	return item.Purchase(memberID, renew)
}
