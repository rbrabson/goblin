package shop

import (
	"testing"

	"github.com/joho/godotenv"

	"github.com/rbrabson/goblin/database/mongo"
	log "github.com/sirupsen/logrus"
)

const (
	GUILD_ID = "12345"
)

var (
	testShop  *Shop
	purchases []*Purchase = make([]*Purchase, 0)
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

func TestGetShopItem(t *testing.T) {
	setup(t)
	defer teardown()

	item := getShopItem(GUILD_ID, "test_item_1", "role")
	if item == nil {
		t.Error("GetShopItem failed to returned n existing item")
	}

	item = testShop.GetShopItem("test_item_1", "role")
	if item == nil {
		t.Error("GetShopItem failed to returned n existing item")
	}
}

func TestRemoveShopItem(t *testing.T) {
	setup(t)
	defer teardown()

	item := testShop.GetShopItem("test_item_1", "role")
	if item == nil {
		t.Error("GetShopItem failed to returned n existing item")
		return
	}

	err := testShop.removeShopItem(item.Name, item.Type)
	if err != nil {
		t.Error("failed to remove shop item, error:")
	}
}

func TestUpdateShopItem(t *testing.T) {
	setup(t)
	defer teardown()

	item := testShop.GetShopItem("test_item_1", "role")
	if item == nil {
		t.Error("GetShopItem failed to returned n existing item")
	}

	err := item.update("test_item_1", "description of test Item 1", "role", 200, "", false)
	if err != nil {
		t.Error("failed to update shop item, error:")
	}

	item = testShop.GetShopItem("test_item_1", "role")
	if item == nil {
		t.Error("GetShopItem failed to returned n existing item")
		return
	}
	if item.Price != 200 {
		t.Error("failed to update shop item price")
	}
}

func setup(t *testing.T) {
	var err error

	testShop = GetShop(GUILD_ID)

	_, err = testShop.addShopItem("test_item_1", "description of test Item 1", "role", 100, "", false)
	if err != nil {
		t.Fatal(err)
	}
	_, err = testShop.addShopItem("test_item_2", "description of test_item_2", "role", 100, "", false)
	if err != nil {
		t.Fatal(err)
	}
	_, err = testShop.addShopItem("test_item_3", "description of test_item_3", "role", 100, "", false)
	if err != nil {
		t.Fatal(err)
	}
}

func teardown() {
	log.Infof("teardown: deleting %d items", len(testShop.Items))
	for _, item := range testShop.Items {
		err := deleteShopItem(item)
		if err != nil {
			log.Error(err)
		}
	}
	log.Infof("teardown: deleting %d purchases", len(purchases))
	for _, purchase := range purchases {
		err := deletePurchase(purchase)
		if err != nil {
			log.Error(err)
		}
	}
}
