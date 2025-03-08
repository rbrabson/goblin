package shop

import (
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	SHOP_ITEM_COLLECTION = "shop_items"
	PURCHASE_COLLECTION  = "shop_purchases"
)

// readShopItems reads all the shop items for the given guild.
func readShopItems(guildID string) ([]*ShopItem, error) {
	log.Trace("--> shop.readShopItems")
	defer log.Trace("<-- shop.readShopItems")

	filter := bson.M{"guild_id": guildID}
	sortBy := bson.M{"name": 1}
	var items []*ShopItem
	err := db.FindMany(SHOP_ITEM_COLLECTION, filter, &items, sortBy, 0)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID}).Error("unable to read shop items from the database")
		return nil, err
	}
	log.WithFields(log.Fields{"guild": guildID, "count": len(items)}).Debug("read shop items from the database")

	return items, nil
}

// readShopItem reads the shop item with the given name and type for the given guild.
func readShopItem(guildID string, name string, itemType string) (*ShopItem, error) {
	log.Trace("--> shop.readShopItem")
	defer log.Trace("<-- shop.readShopItem")

	filter := bson.D{{Key: "guild_id", Value: guildID}, {Key: "name", Value: name}, {Key: "type", Value: itemType}}
	var item *ShopItem
	err := db.FindOne(SHOP_ITEM_COLLECTION, filter, &item)
	if err != nil {
		log.WithFields(log.Fields{"filter": filter, "error": err}).Error("unable to read shop item from the database")
		return nil, err
	}
	log.WithFields(log.Fields{"guild": guildID, "name": item.Name, "type": item.Type}).Debug("read shop item from the database")

	return item, nil
}

// writeShopItem writes the shop item to the database.
func writeShopItem(item *ShopItem) error {
	log.Trace("--> shop.writeShopItem")
	defer log.Trace("<-- shop.writeShopItem")

	var filter bson.D
	if item.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: item.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: item.GuildID}, {Key: "name", Value: item.Name}, {Key: "type", Value: item.Type}}
	}
	err := db.UpdateOrInsert(SHOP_ITEM_COLLECTION, filter, item)
	if err != nil {
		log.WithFields(log.Fields{"item": item, "error": err}).Error("unable to save shop item to the database")
		return err
	}
	log.WithFields(log.Fields{"item": item, "filter": filter}).Debug("write the shop item to the database")

	return nil
}

// deleteShopItem deletes the shop item from the database.
func deleteShopItem(item *ShopItem) error {
	log.Trace("--> shop.deleteShopItem")
	defer log.Trace("<-- shop.deleteShopItem")

	var filter bson.D
	if item.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: item.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: item.GuildID}, {Key: "name", Value: item.Name}, {Key: "type", Value: item.Type}}
	}
	err := db.Delete(SHOP_ITEM_COLLECTION, filter)
	if err != nil {
		log.WithFields(log.Fields{"item": item, "error": err}).Error("unable to delete shop item from the database")
		return err
	}
	log.WithFields(log.Fields{"item": item, "filter": filter}).Debug("delete the shop item from the database")

	return nil
}

// readPurchases reads all the purchases for the member in the given guild.
func readPurchases(guildID string, memberID string) ([]*Purchase, error) {
	log.Trace("--> shop.readPurchases")
	defer log.Trace("<-- shop.readPurchases")

	filter := bson.M{"guildID": guildID, "member_id": memberID}
	sortBy := bson.M{"item.name": 1}
	var items []*Purchase
	err := db.FindMany(PURCHASE_COLLECTION, filter, &items, sortBy, 0)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "member_id": memberID}).Error("unable to read purchases from the database")
		return nil, err
	}
	log.WithFields(log.Fields{"guildID": guildID, "memberID": memberID, "count": len(items)}).Debug("read shop items from the database")

	return items, nil
}

// readPurchase reads the purchase with the given name and type for the given guild.
func readPurchase(guildID string, memberID string, itemName string, itemType string) (*Purchase, error) {
	log.Trace("--> shop.readPurchases")
	defer log.Trace("<-- shop.readPurchases")

	filter := bson.M{"guildID": guildID, "member_id": memberID, "item.name": itemName, "item.type": itemType}
	var item *Purchase
	err := db.FindOne(PURCHASE_COLLECTION, filter, &item)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "member_id": memberID, "item.name": itemName, "item.type": itemType}).Error("unable to read purchase from the database")
		return nil, err
	}
	log.WithFields(log.Fields{"guildID": guildID, "memberID": memberID, "item.name": itemName, "item.type": itemType}).Debug("read shop item from the database")

	return item, nil
}

// writePurchases writes the purchase to the database.
func writePurchase(item *Purchase) error {
	log.Trace("--> shop.writeShopItem")
	defer log.Trace("<-- shop.writeShopItem")

	var filter bson.D
	if item.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: item.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: item.GuildID}, {Key: "member_id", Value: item.MemberID}, {Key: "item.name", Value: item.Item.Name}, {Key: "item.type", Value: item.Item.Type}}
	}
	err := db.UpdateOrInsert(SHOP_ITEM_COLLECTION, filter, item)
	if err != nil {
		log.WithFields(log.Fields{"item": item, "error": err}).Error("unable to write purchase to the database")
		return err
	}
	log.WithFields(log.Fields{"item": item}).Debug("write purchase to the database")

	return nil
}

// deletePurchase deletes the purchase from the database.
func deletePurchase(purchase *Purchase) error {
	log.Trace("--> shop.deletePurchase")
	defer log.Trace("<-- shop.deletePurchase")

	var filter bson.D
	if purchase.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: purchase.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: purchase.GuildID}, {Key: "member_id", Value: purchase.MemberID}, {Key: "item.name", Value: purchase.Item.Name}, {Key: "item.type", Value: purchase.Item.Type}}
	}
	err := db.Delete(SHOP_ITEM_COLLECTION, filter)
	if err != nil {
		log.WithFields(log.Fields{"purchase": purchase, "error": err}).Error("unable to delete purchasefrom the database")
		return err
	}
	log.WithFields(log.Fields{"purchase": purchase}).Debug("delete the purchasefrom the database")

	return nil
}
