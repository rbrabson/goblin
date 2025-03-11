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

// readAllPurchases reads all the purchases from the database that match the input filter
func readAllPurchases(filter interface{}) ([]*Purchase, error) {
	log.Trace("--> shop.readAllPurchases")
	defer log.Trace("<-- shop.readAllPurchases")

	var items []*Purchase
	err := db.FindMany(PURCHASE_COLLECTION, filter, &items, bson.D{}, 0)
	if err != nil {
		log.WithFields(log.Fields{"filter": filter}).Error("unable to read purchases from the database")
		return nil, err
	}
	log.WithFields(log.Fields{"count": len(items)}).Debug("read purchases from the database")

	return items, nil
}

// readPurchases reads all the purchases for the member in the given guild.
func readPurchases(guildID string, memberID string) ([]*Purchase, error) {
	log.Trace("--> shop.readPurchases")
	defer log.Trace("<-- shop.readPurchases")

	filter := bson.M{"guild_id": guildID, "member_id": memberID}
	sortBy := bson.M{"name": 1}
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

	filter := bson.D{{Key: "guild_id", Value: guildID}, {Key: "member_id", Value: memberID}, {Key: "name", Value: itemName}, {Key: "type", Value: itemType}}
	var item Purchase
	err := db.FindOne(PURCHASE_COLLECTION, filter, &item)
	if err != nil {
		log.WithFields(log.Fields{"filter": filter}).Debug("unable to read purchase from the database")
		return nil, err
	}
	log.WithFields(log.Fields{"guildID": guildID, "memberID": memberID, "name": itemName, "type": itemType}).Debug("read shop item from the database")

	return &item, nil
}

// writePurchases writes the purchase to the database.
func writePurchase(item *Purchase) error {
	log.Trace("--> shop.writeShopItem")
	defer log.Trace("<-- shop.writeShopItem")

	var filter bson.D
	if item.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: item.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: item.Item.GuildID}, {Key: "member_id", Value: item.MemberID}, {Key: "name", Value: item.Item.Name}, {Key: "type", Value: item.Item.Type}}
	}
	err := db.UpdateOrInsert(PURCHASE_COLLECTION, filter, item)
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
	if purchase.Item.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: purchase.Item.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: purchase.Item.GuildID}, {Key: "member_id", Value: purchase.MemberID}, {Key: "name", Value: purchase.Item.Name}, {Key: "type", Value: purchase.Item.Type}}
	}
	err := db.Delete(PURCHASE_COLLECTION, filter)
	if err != nil {
		log.WithFields(log.Fields{"purchase": purchase, "error": err}).Error("unable to delete purchasefrom the database")
		return err
	}
	log.WithFields(log.Fields{"purchase": purchase}).Debug("delete the purchase from the database")

	return nil
}
