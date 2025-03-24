package shop

import (
	"fmt"
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

// addShopItem adds a new item to the shop. If the item already exists, an error is returned.
func (s *Shop) addShopItem(name string, description string, itemType string, price int, duration string, renewable bool) (*ShopItem, error) {
	item := s.GetShopItem(name, itemType)
	if item != nil {
		return nil, fmt.Errorf("item already exists")
	}

	item = newShopItem(s.GuildID, name, description, itemType, price, duration, renewable)
	if item == nil {
		log.WithFields(log.Fields{"guild": s.GuildID, "name": name, "type": itemType}).Error("unable to write shop item to the database")
		return nil, fmt.Errorf("unable to add item")
	}
	s.Items = append(s.Items, item)

	log.WithFields(log.Fields{"guild": item.GuildID, "name": item.Name, "type": item.Type}).Info("shop item added")
	return item, nil
}

// removeShopItem removes an item from the shop. If the item does not exist, an error is returned.
func (s *Shop) removeShopItem(name string, itemType string) error {
	item := s.GetShopItem(name, itemType)
	if item == nil {
		return fmt.Errorf("item does not exist")
	}

	err := deleteShopItem(item)
	if err != nil {
		return fmt.Errorf("unable to remove item")
	}

	log.WithFields(log.Fields{"guild": item.GuildID, "name": item.Name, "type": item.Type}).Info("shop item removed")
	return nil
}
