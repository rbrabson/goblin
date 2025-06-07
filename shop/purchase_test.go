package shop

import (
	"log/slog"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/database/mongo"
)

const (
	MemberId = "67890"
)

func init() {
	err := godotenv.Load("../.env_test")
	if err != nil {
		slog.Error("Error loading .env file")
		os.Exit(1)
	}
	db = mongo.NewDatabase()
	bank.SetDB(db)
	testShop = GetShop(GuildId)
}

func TestNewPurchase(t *testing.T) {
	setup(t)
	defer teardown()

	item := testShop.GetShopItem("test_item_1", "role")
	purchase, err := PurchaseItem(GuildId, MemberId, item, PURCHASED, false)
	if err != nil {
		t.Errorf("NewPurchase failed to create a new purchase, error: %s", err)
	}
	purchases = append(purchases, purchase)
}

func TestGetAllPurchases(t *testing.T) {
	setup(t)
	defer teardown()

	item1 := testShop.GetShopItem("test_item_1", "role")
	purchase, err := PurchaseItem(GuildId, MemberId, item1, PURCHASED, false)
	if err != nil {
		t.Errorf("NewPurchase failed to create a new purchase, error: %s", err)
		return
	}
	purchases = append(purchases, purchase)
	slog.Debug("purchases",
		slog.Any("purchases", purchases),
	)

	item2 := testShop.GetShopItem("test_item_2", "role")
	purchase, err = PurchaseItem(GuildId, MemberId, item2, PURCHASED, false)
	if err != nil {
		t.Errorf("NewPurchase failed to create a new purchase, error: %s", err)
		return
	}
	purchases = append(purchases, purchase)
	slog.Debug("purchases",
		slog.Any("purchases", purchases),
	)

	locPurchases := GetAllPurchases(GuildId, MemberId)
	slog.Info("Purchases from DB", slog.Any("purcahses", locPurchases))
	if len(locPurchases) != 2 {
		t.Errorf("GetAllPurchases failed to return all purchases, expected 2, got %d", len(locPurchases))
		t.Errorf("purchases: %v", locPurchases)
		return
	}
}

func TestUpdatePurchase(t *testing.T) {
	setup(t)
	defer teardown()

	item := testShop.GetShopItem("test_item_1", "role")
	purchase, err := PurchaseItem(GuildId, MemberId, item, PURCHASED, false)
	if err != nil {
		t.Errorf("NewPurchase failed to create a new purchase, error: %s", err)
		return
	}
	purchases = append(purchases, purchase)

	err = purchase.Update(true)
	if err != nil {
		t.Errorf("UpdatePurchase failed to update the purchase, error: %s", err)
		return
	}
	purchase, err = readPurchase(purchase.GuildID, purchase.MemberID, purchase.Item.Name, purchase.Item.Type)
	if err != nil {
		t.Errorf("UpdatePurchase failed to read the purchase, error: %s", err)
		return
	}
	if purchase.AutoRenew != true {
		t.Errorf("UpdatePurchase failed to update the purchase, expected true, got %v", purchase.AutoRenew)
		return
	}
}
