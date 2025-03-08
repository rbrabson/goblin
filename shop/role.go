package shop

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	ROLE_TYPE = "role"
)

// Role is a custom role that a member can purchase in the shop.
type Role ShopItem

// GetRole returns the role with the given name.
func GetRole(guildID string, name string) (*Role, error) {
	log.Trace("--> shop.GetRole")
	defer log.Trace("<-- shop.GetRole")

	item := GetShopItem(guildID, name, ROLE_TYPE)
	if item == nil {
		log.WithFields(log.Fields{"guild": guildID, "item": name}).Error("role not found")
		return nil, fmt.Errorf("role %s not found", name)
	}

	role := Role(*item)
	return &role, nil
}

// NewRole returns a new role that can be purchased in the shop
func NewRole(guildID string, name string, price int, description string, duration time.Duration, renewable bool) (*ShopItem, error) {
	log.Trace("--> shop.NewRole")
	defer log.Trace("<-- shop.NewRole")

	role := NewShopItem(guildID, name, description, ROLE_TYPE, price, duration, renewable)
	err := writeShopItem(role)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "item": name}).Error("unable to create role")
		return nil, err
	}
	return role, nil
}

// Purchase buys the role for the given guild and member.
// If the member already owns the role, an error is returned.
func (r *Role) Purchase(memberID string, autoRenew bool) (*Purchase, error) {
	log.Trace("--> shop.Role.Buy")
	defer log.Trace("<-- shop.Role.Buy")

	purchase := GetPurchase(r.GuildID, memberID, r.Name, ROLE_TYPE)
	if purchase != nil {
		log.WithFields(log.Fields{"guild": r.GuildID, "member": memberID, "item": r.Name}).Info("member already owns this role")
		return nil, fmt.Errorf("member already owns role %s", r.Name)
	}

	var err error
	shopItem := ShopItem(*r)
	purchase, err = NewPurchase(r.GuildID, memberID, &shopItem, autoRenew)
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{"guild": r.GuildID, "member": memberID, "item": r.Name}).Info("purchase created successfully")

	return purchase, nil
}
