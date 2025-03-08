package shop

import (
	"testing"

	"github.com/joho/godotenv"
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
	testShop = GetShop(GUILD_ID)
}

func TestGetAllPurchsableItems(t *testing.T) {
	setup(t)
	defer teardown(t)

	items := GetAllPurchasableItems(GUILD_ID)
	if len(items) != 3 {
		t.Errorf("GetAllPurchasableItems failed to return all items, expected 2, got %d", len(items))
	}
}

func TestNewPurchase(t *testing.T) {
	setup(t)
	defer teardown(t)

	item := testShop.GetShopItem("test_item_1", "role")
	purchase, err := NewPurchase(GUILD_ID, MEMBER_ID, item, false)
	if err != nil {
		t.Errorf("NewPurchase failed to create a new purchase, error: %s", err)
	}
	purchases = append(purchases, purchase)
}
