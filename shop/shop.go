package shop

import (
	"cmp"
	"log/slog"
	"slices"

	"github.com/rbrabson/goblin/internal/logger"
)

var (
	sslog = logger.GetLogger()
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
		sslog.Error("unable to read shop items from the database",
			slog.String("guildID", guildID),
			slog.Any("error", err),
		)
		shop.Items = make([]*ShopItem, 0)
	}

	shopItemCmp := func(a, b *ShopItem) int {
		return cmp.Or(
			cmp.Compare(a.Type, b.Type),
			cmp.Compare(a.Name, b.Name),
		)
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
