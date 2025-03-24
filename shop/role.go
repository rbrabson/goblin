package shop

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/guild"
	log "github.com/sirupsen/logrus"
)

const (
	ROLE = "role"
)

// Role represents a role item in the shop.
type Role ShopItem

// GetRole retrieves a role from the shop by its name for a specific guild.
func GetRole(guildID string, name string) *Role {
	item := getShopItem(guildID, name, ROLE)
	role := Role(*item)
	return &role
}

// NewRole creates a new role for the shop.
func NewRole(guildID string, name string, description string, price int, duration string, autoRenewable bool) *Role {
	item := newShopItem(guildID, name, description, ROLE, price, duration, autoRenewable)
	role := (*Role)(item)
	return role
}

// Update updates the role's properties in the shop.
func (r *Role) Update(name string, description string, price int, duration string, autoRenewable bool) error {
	item := (*ShopItem)(r)
	return item.update(name, description, ROLE, price, duration, autoRenewable)
}

// Purchase allows a member to purchase the role from the shop.
func (r *Role) Purchase(memberID string, renew bool) (*Purchase, error) {
	item := ShopItem(*r)
	return item.purchase(memberID, renew)
}

// AddToShop adds the role to the shop. If the role already exists, an error is returned.
func (r *Role) AddToShop(s *Shop) error {
	item := (*ShopItem)(r)
	return item.addToShop(s)
}

// RemoveFromShop removes the role from the shop. If the role does not exist, an error is returned.
func (r *Role) RemoveFromShop(s *Shop) error {
	item := (*ShopItem)(r)
	return item.removeFromShop(s)
}

// rolePurchaseChecks performs checks to see if a role can be purchased.
func rolePurchaseChecks(s *discordgo.Session, i *discordgo.InteractionCreate, roleName string) error {
	// Verify the role exists on the server
	guildRole := guild.GetGuildRole(s, i.GuildID, roleName)
	if guildRole == nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Error("role not found on server")
		return fmt.Errorf("role `%s` not found on the server", roleName)
	}

	// Make sure the member doesn't already have the role
	if guild.MemberHasRole(s, i.GuildID, i.Member.User.ID, guildRole) {
		return fmt.Errorf("you already have the `%s` role", roleName)
	}

	// Make sure the role is still available in the shop
	shopItem := getShopItem(i.GuildID, roleName, ROLE)
	if shopItem == nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Error("failed to read role from shop")
		return fmt.Errorf("role `%s` not found in the shop", roleName)
	}

	// Make common checks for all purchases
	err := purchaseChecks(i.GuildID, i.Member.User.ID, ROLE, roleName)
	if err != nil {
		return err
	}

	return nil
}
