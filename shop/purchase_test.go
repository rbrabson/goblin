package shop

import (
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

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

func TestPurchaseItemInsufficientFunds(t *testing.T) {
	setup(t)
	defer teardown()

	// Get the account and set balance to 0
	account := bank.GetAccount(GuildId, MemberId)
	initialBalance := account.CurrentBalance

	// Set balance to 0
	if err := account.SetBalance(0); err != nil {
		t.Errorf("Failed to set balance to 0: %s", err)
		return
	}

	// Try to purchase an item with insufficient funds
	item := testShop.GetShopItem("test_item_1", "role")
	purchase, err := PurchaseItem(GuildId, MemberId, item, PURCHASED, false)

	// Verify that the purchase failed due to insufficient funds
	if err == nil {
		t.Errorf("Expected PurchaseItem to fail due to insufficient funds, but it succeeded")
		purchases = append(purchases, purchase)
		return
	}

	// Verify that the error message contains "insufficient funds"
	if !strings.Contains(strings.ToLower(err.Error()), "insufficient funds") {
		t.Errorf("Expected error message to contain 'insufficient funds', got: %s", err.Error())
	}

	// Restore the initial balance
	if err := account.SetBalance(initialBalance); err != nil {
		t.Errorf("Failed to restore initial balance: %s", err)
	}
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

func TestHasExpired(t *testing.T) {
	setup(t)
	defer teardown()

	// Create a purchase with an expiration time in the past
	item := testShop.GetShopItem("test_item_1", "role")
	purchase, err := PurchaseItem(GuildId, MemberId, item, PURCHASED, false)
	if err != nil {
		t.Errorf("PurchaseItem failed to create a new purchase, error: %s", err)
		return
	}
	purchases = append(purchases, purchase)

	// Set expiration time to 1 hour ago
	purchase.ExpiresOn = time.Now().Add(-1 * time.Hour)
	err = writePurchase(purchase)
	if err != nil {
		t.Errorf("Failed to update purchase expiration time: %s", err)
		return
	}

	// Verify the purchase has expired
	expired := purchase.HasExpired()
	if !expired {
		t.Errorf("Expected purchase to be expired, but it wasn't")
	}

	// Create a purchase with an expiration time in the future
	item = testShop.GetShopItem("test_item_2", "role")
	purchase, err = PurchaseItem(GuildId, MemberId, item, PURCHASED, false)
	if err != nil {
		t.Errorf("PurchaseItem failed to create a new purchase, error: %s", err)
		return
	}
	purchases = append(purchases, purchase)

	// Set expiration time to 1 hour in the future
	purchase.ExpiresOn = time.Now().Add(1 * time.Hour)
	err = writePurchase(purchase)
	if err != nil {
		t.Errorf("Failed to update purchase expiration time: %s", err)
		return
	}

	// Verify the purchase has not expired
	expired = purchase.HasExpired()
	if expired {
		t.Errorf("Expected purchase to not be expired, but it was")
	}
}

func TestCheckForExpiredPurchases(t *testing.T) {
	setup(t)
	defer teardown()

	// Create a purchase with an expiration time in the past
	item := testShop.GetShopItem("test_item_1", "role")
	purchase, err := PurchaseItem(GuildId, MemberId, item, PURCHASED, false)
	if err != nil {
		t.Errorf("PurchaseItem failed to create a new purchase, error: %s", err)
		return
	}
	purchases = append(purchases, purchase)

	// Set expiration time to 1 hour ago
	purchase.ExpiresOn = time.Now().Add(-1 * time.Hour)
	err = writePurchase(purchase)
	if err != nil {
		t.Errorf("Failed to update purchase expiration time: %s", err)
		return
	}

	// Run the check for expired purchases
	checkForExpiredPurchases()

	// Verify the purchase is marked as expired
	purchase, err = readPurchase(purchase.GuildID, purchase.MemberID, purchase.Item.Name, purchase.Item.Type)
	if err != nil {
		t.Errorf("Failed to read purchase after checking for expired purchases: %s", err)
		return
	}

	if !purchase.IsExpired {
		t.Errorf("Purchase should be marked as expired after checkForExpiredPurchases")
	}
}

func TestReturnPurchase(t *testing.T) {
	setup(t)
	defer teardown()

	// Get the initial balance
	account := bank.GetAccount(GuildId, MemberId)
	initialBalance := account.CurrentBalance

	// Purchase an item
	item := testShop.GetShopItem("test_item_1", "role")
	purchase, err := PurchaseItem(GuildId, MemberId, item, PURCHASED, false)
	if err != nil {
		t.Errorf("PurchaseItem failed to create a new purchase, error: %s", err)
		return
	}
	purchases = append(purchases, purchase)

	// Verify the balance was reduced by the item price
	account = bank.GetAccount(GuildId, MemberId)
	if account.CurrentBalance != initialBalance-item.Price {
		t.Errorf("Expected balance to be reduced by %d, got %d", item.Price, account.CurrentBalance)
		return
	}

	// Return the purchase
	err = purchase.Return()
	if err != nil {
		t.Errorf("Return failed to return the purchase, error: %s", err)
		return
	}

	// Verify the balance was restored
	account = bank.GetAccount(GuildId, MemberId)
	if account.CurrentBalance != initialBalance {
		t.Errorf("Expected balance to be restored to %d, got %d", initialBalance, account.CurrentBalance)
		return
	}

	// Verify the purchase was deleted
	purchases := GetAllPurchases(GuildId, MemberId)
	for _, p := range purchases {
		if p.Item.Name == item.Name && p.Item.Type == item.Type {
			t.Errorf("Purchase was not deleted after return")
			return
		}
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
