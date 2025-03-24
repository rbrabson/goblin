package shop

import (
	"slices"

	log "github.com/sirupsen/logrus"
)

// The shop for a guild. The shop contains all items available for purchase.
type Shop struct {
	GuildID string      // Guild (server) for the shop
	Items   []*ShopItem // All items available in the shop
}

// GetShop returns the shop for the guild.
func GetShop(guildID string) *Shop {
	var err error

	shop := &Shop{
		GuildID: guildID,
	}

	shop.Items, err = readShopItems(guildID)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "error": err}).Error("unable to read shop items from the database")
		shop.Items = make([]*ShopItem, 0)
	}

	shopItemCmp := func(a, b *ShopItem) int {
		if a.Type < b.Type {
			return -1
		}
		if a.Type > b.Type {
			return 1
		}
		if a.Name < b.Name {
			return -1
		}
		if a.Name > b.Name {
			return 1
		}
		return 0
	}
	slices.SortFunc(shop.Items, shopItemCmp)

	return shop
}

// GetShopItems finds an item in the shop. If the item does not exist then nil is returned.
func (s *Shop) GetShopItem(name string, itemType string) *ShopItem {
	for _, item := range s.Items {
		if item.Name == name && item.Type == itemType {
			return item
		}
	}

	return nil
}
