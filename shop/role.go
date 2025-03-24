package shop

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/bank"
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

// rolePurchaseChecks performs checks to see if a role can be purchased.
func rolePurchaseChecks(s *discordgo.Session, i *discordgo.InteractionCreate, roleName string) error {
	log.Trace("--> shop.rolePurchaseChecks")
	defer log.Trace("<-- shop.rolePurchaseChecks")

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
	shopItem := GetShopItem(i.GuildID, roleName, ROLE)
	if shopItem == nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Error("failed to read role from shop")
		return fmt.Errorf("role `%s` not found in the shop", roleName)
	}

	// Make sure the role hasn't already been purchased
	purchase, _ := readPurchase(i.GuildID, i.Member.User.ID, roleName, ROLE)
	if purchase != nil && !purchase.IsExpired {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Debug("role already purchased")
		return fmt.Errorf("you have already purchased role `%s`", roleName)
	}

	// Make sure the member has sufficient funds to purchase the role
	bankAccount := bank.GetAccount(i.GuildID, i.Member.User.ID)
	if bankAccount.CurrentBalance < shopItem.Price {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "memberID": i.Member.User.ID}).Debug("insufficient funds")
		return fmt.Errorf("you do not have enough credits to purchase the `%s` role", roleName)
	}
	return nil
}
