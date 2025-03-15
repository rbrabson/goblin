package shop

import (
	"testing"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/database/mongo"
	log "github.com/sirupsen/logrus"
)

const (
	MEMBER_ID = "67890"
)

func init() {
	err := godotenv.Load("../.env_test")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.SetLevel(log.DebugLevel)
	db = mongo.NewDatabase()
	bank.SetDB(db)
	testShop = GetShop(GUILD_ID)
}

func TestNewPurchase(t *testing.T) {
	setup(t)
	defer teardown()

	item := testShop.GetShopItem("test_item_1", "role")
	purchase, err := PurchaseItem(GUILD_ID, MEMBER_ID, item, false)
	if err != nil {
		t.Errorf("NewPurchase failed to create a new purchase, error: %s", err)
	}
	purchases = append(purchases, purchase)
}

func TestGetAllPurchases(t *testing.T) {
	setup(t)
	defer teardown()

	item1 := testShop.GetShopItem("test_item_1", "role")
	purchase, err := PurchaseItem(GUILD_ID, MEMBER_ID, item1, false)
	if err != nil {
		t.Errorf("NewPurchase failed to create a new purchase, error: %s", err)
		return
	}
	purchases = append(purchases, purchase)
	log.WithField("purchases", purchases).Debug("purchases")

	item2 := testShop.GetShopItem("test_item_2", "role")
	purchase, err = PurchaseItem(GUILD_ID, MEMBER_ID, item2, false)
	if err != nil {
		t.Errorf("NewPurchase failed to create a new purchase, error: %s", err)
		return
	}
	purchases = append(purchases, purchase)
	log.WithField("purchases", purchases).Error("purchases")

	locPurchases := GetAllPurchases(GUILD_ID, MEMBER_ID)
	log.Infof("Purchases from DB: %v", locPurchases)
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
	purchase, err := PurchaseItem(GUILD_ID, MEMBER_ID, item, false)
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
