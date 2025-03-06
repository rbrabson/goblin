package shop

const (
	SHOP_ITEM_COLLECTION = "shop_items"
	PURCHASE_COLLECTION  = "shop_purchases"
)

func readAllShopItems(guildID string) ([]*ShopItem, error) {
	return nil, nil
}

func readShopIem(guildID string, name string, itemType string) (*ShopItem, error) {
	return nil, nil
}

func writeShopItem(item *ShopItem) error {
	return nil
}

func deleteShopItem(item *ShopItem) error {
	return nil
}

func readAllPurchases(guildID string, memberID string) ([]*Purchase, error) {
	return nil, nil
}

func writePurchase(purchase *Purchase) error {
	return nil
}
