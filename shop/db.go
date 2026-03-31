package shop

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ConfigCollection   = "shop_configs"
	ShopItemCollection = "shop_items"
	PurchaseCollection = "shop_purchases"
	MemberCollection   = "shop_members"
)

// readConfig reads the configuration from the database. If the config does not exist, it returns nil.
func readConfig(guildID string) (*Config, error) {
	filter := bson.M{"guild_id": guildID}
	var config *Config
	err := db.FindOne(ConfigCollection, filter, &config)
	if err != nil {
		slog.Error("unable to read shop config from the database", "guildID", guildID, "filter", filter, "error", err)
		return nil, err
	}
	slog.Debug("read shop config from the database", "guildID", guildID)

	return config, nil
}

// writeConfig writes the configuration to the database.
func writeConfig(config *Config) error {
	var filter bson.D
	if config.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: config.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: config.GuildID}}
	}
	err := db.UpdateOrInsert(ConfigCollection, filter, config)
	if err != nil {
		slog.Error("unable to write shop config to the database", "guildID", config.GuildID, "filter", filter, "error", err)
		return err
	}
	slog.Debug("write shop config to the database", "guildID", config.GuildID)

	return nil
}

// readShopItems reads all the shop items for the given guild.
func readShopItems(guildID string) ([]*ShopItem, error) {
	filter := bson.M{"guild_id": guildID}
	sortBy := bson.M{"name": 1}
	var items []*ShopItem
	err := db.FindMany(ShopItemCollection, filter, &items, sortBy, 0)
	if err != nil {
		slog.Error("unable to read shop items from the database", "guildID", guildID, "filter", filter, "error", err)
		return nil, err
	}
	slog.Debug("read shop items from the database", "guildID", guildID, "count", len(items))

	return items, nil
}

// readShopItem reads the shop item with the given name and type for the given guild.
func readShopItem(guildID string, name string, itemType string) (*ShopItem, error) {
	filter := bson.D{{Key: "guild_id", Value: guildID}, {Key: "name", Value: name}, {Key: "type", Value: itemType}}
	var item *ShopItem
	err := db.FindOne(ShopItemCollection, filter, &item)
	if err != nil {
		slog.Error("unable to read shop item from the database", "guildID", guildID, "filter", filter, "error", err)
		return nil, err
	}
	slog.Debug("read shop item from the database", "guildID", guildID, "name", name, "type", itemType)

	return item, nil
}

// writeShopItem writes the shop item to the database.
func writeShopItem(item *ShopItem) error {
	var filter bson.D
	if item.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: item.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: item.GuildID}, {Key: "name", Value: item.Name}, {Key: "type", Value: item.Type}}
	}
	err := db.UpdateOrInsert(ShopItemCollection, filter, item)
	if err != nil {
		slog.Error("unable to save shop item to the database", "guildID", item.GuildID, "filter", filter, "error", err)
		return err
	}
	slog.Debug("write the shop item to the database", "guildID", item.GuildID, "name", item.Name, "type", item.Type)

	return nil
}

// deleteShopItem deletes the shop item from the database.
func deleteShopItem(item *ShopItem) error {
	var filter bson.D
	if item.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: item.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: item.GuildID}, {Key: "name", Value: item.Name}, {Key: "type", Value: item.Type}}
	}
	err := db.Delete(ShopItemCollection, filter)
	if err != nil {
		slog.Error("unable to delete shop item from the database", "guildID", item.GuildID, "filter", filter, "error", err)
		return err
	}
	slog.Debug("delete the shop item from the database", "guildID", item.GuildID, "name", item.Name, "filter", filter)

	return nil
}

// readAllPurchases reads all the purchases from the database that match the input filter
func readAllPurchases(filter interface{}) ([]*Purchase, error) {
	var items []*Purchase
	err := db.FindMany(PurchaseCollection, filter, &items, bson.D{}, 0)
	if err != nil {
		slog.Error("unable to read all purchases from the database", "filter", filter, "error", err)
		return nil, err
	}

	return items, nil
}

// readPurchases reads all the purchases for the member in the given guild.
func readPurchases(guildID string, memberID string) ([]*Purchase, error) {
	filter := bson.M{"guild_id": guildID, "member_id": memberID}
	sortBy := bson.M{"name": 1}
	var items []*Purchase
	err := db.FindMany(PurchaseCollection, filter, &items, sortBy, 0)
	if err != nil {
		slog.Error("unable to read purchases from the database", "guildID", guildID, "memberID", memberID, "filter", filter, "error", err)
		return nil, err
	}

	return items, nil
}

// readPurchase reads the purchase with the given name and type for the given guild.
func readPurchase(guildID string, memberID string, itemName string, itemType string) (*Purchase, error) {
	filter := bson.D{{Key: "guild_id", Value: guildID}, {Key: "member_id", Value: memberID}, {Key: "name", Value: itemName}, {Key: "type", Value: itemType}, {Key: "is_expired", Value: false}}
	var item Purchase
	err := db.FindOne(PurchaseCollection, filter, &item)
	if err != nil {
		slog.Debug("unable to read purchase from the database", "filter", filter, "error", err)
		return nil, err
	}
	slog.Debug("read shop item from the database", "guildID", guildID, "filter", filter)

	return &item, nil
}

// writePurchases writes the purchase to the database.
func writePurchase(item *Purchase) error {
	var filter bson.D
	if item.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: item.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: item.Item.GuildID}, {Key: "member_id", Value: item.MemberID}, {Key: "name", Value: item.Item.Name}, {Key: "type", Value: item.Item.Type}, {Key: "is_expired", Value: false}}
	}
	err := db.UpdateOrInsert(PurchaseCollection, filter, item)
	if err != nil {
		slog.Error("unable to write purchase to the database", "guildID", item.Item.GuildID, "filter", filter, "item", item, "error", err)
		return err
	}
	slog.Debug("write purchase to the database", "guildID", item.Item.GuildID, "filter", filter, "item", item)

	return nil
}

// deletePurchase deletes the purchase from the database.
func deletePurchase(purchase *Purchase) error {
	var filter bson.D
	if purchase.Item.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: purchase.Item.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: purchase.Item.GuildID}, {Key: "member_id", Value: purchase.MemberID}, {Key: "name", Value: purchase.Item.Name}, {Key: "type", Value: purchase.Item.Type}}
	}
	err := db.Delete(PurchaseCollection, filter)
	if err != nil {
		slog.Error("unable to delete purchase from the database", "guildID", purchase.Item.GuildID, "filter", filter, "error", err)
		return err
	}
	slog.Debug("delete the purchase from the database", "guildID", purchase.Item.GuildID, "filter", filter)

	return nil
}

// readMember reads the member from the database.
func readMember(guildID string, memberID string) (*Member, error) {
	filter := bson.D{{Key: "guild_id", Value: guildID}, {Key: "member_id", Value: memberID}}
	var member *Member
	err := db.FindOne(MemberCollection, filter, &member)
	if err != nil {
		slog.Debug("unable to read shop member from the database", "guildID", guildID, "memberID", memberID, "error", err)
		return nil, err
	}
	slog.Debug("read shop member from the database", "guildID", guildID, "memberID", memberID)

	return member, nil
}

// writeMember writes the member to the database.
func writeMember(member *Member) error {
	var filter bson.D
	if member.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: member.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: member.GuildID}, {Key: "member_id", Value: member.MemberID}}
	}
	err := db.UpdateOrInsert(MemberCollection, filter, member)
	if err != nil {
		slog.Error("unable to save shop member to the database", "guildID", member.GuildID, "memberID", member.MemberID, "filter", filter, "error", err)
		return err
	}
	slog.Info("write the shop member to the database", "guildID", member.GuildID, "memberID", member.MemberID)

	return nil
}

// deleteShopItem deletes the shop item from the database.
func deleteMember(member *Member) error {
	var filter bson.D
	if member.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: member.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: member.GuildID}, {Key: "member_id", Value: member.MemberID}}
	}
	err := db.Delete(MemberCollection, filter)
	if err != nil {
		slog.Error("unable to delete shop member from the database", "guildID", member.GuildID, "memberID", member.MemberID, "filter", filter, "error", err)
		return err
	}
	slog.Info("delete the shop member from the database", "guildID", member.GuildID, "memberID", member.MemberID)

	return nil
}

// listMembers lists all the members in the given guild.
func listMembers(guildID string) ([]*Member, error) {
	filter := bson.D{{Key: "guild_id", Value: guildID}}
	var members []*Member
	err := db.FindMany(MemberCollection, filter, &members, bson.D{}, 0)
	if err != nil {
		slog.Error("unable to read shop members from the database", "guildID", guildID, "filter", filter, "error", err)
		return nil, err
	}
	slog.Debug("read shop members from the database", "guildID", guildID, "count", len(members))

	return members, nil
}
