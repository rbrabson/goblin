package shop

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/internal/discmsg"
	"github.com/rbrabson/goblin/internal/disctime"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/text/language"
)

const (
	APPROVED  = "approved"
	DENIED    = "denied"
	PENDING   = "pending"
	PURCHASED = "purchased"
)

// Purchase is a purchase made from the shop.
type Purchase struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID     string             `json:"guild_id" bson:"guild_id"`
	MemberID    string             `json:"member_id" bson:"member_id"`
	Item        *ShopItem          `json:"item" bson:"item,inline"`
	Status      string             `json:"status" bson:"status"`
	PurchasedOn time.Time          `json:"purchased_on" bson:"purchased_on"`
	ExpiresOn   time.Time          `json:"expires_on" bson:"expires_on"`
	AutoRenew   bool               `json:"autoRenew" bson:"autoRenew"`
	IsExpired   bool               `json:"is_expired" bson:"is_expired"`
}

// GetAllRoles returns all the purchases made by a member in the guild.
func GetAllPurchases(guildID string, memberID string) []*Purchase {
	log.Trace("--> shop.GetAllPurchases")
	defer log.Trace("<-- shop.GetAllPurchases")

	purchases, err := readPurchases(guildID, memberID)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "member": memberID, "error": err}).Error("unable to read purchases from the database")
		return nil
	}

	purchaseCmp := func(a, b *Purchase) int {
		// Sort expired purchases to the bottom of the purchases
		if a.HasExpired() && !b.HasExpired() {
			return 1
		}
		if !a.HasExpired() && b.HasExpired() {
			return -1
		}

		// Sort on the basic purchase information
		if a.Item.Type < b.Item.Type {
			return -1
		}
		if a.Item.Type > b.Item.Type {
			return 1
		}
		if a.Item.Name < b.Item.Name {
			return -1
		}
		if a.Item.Name > b.Item.Name {
			return 1
		}
		if a.PurchasedOn.Before(b.PurchasedOn) {
			return -1
		}
		if a.PurchasedOn.After(b.PurchasedOn) {
			return 1
		}
		return 0
	}
	slices.SortFunc(purchases, purchaseCmp)

	return purchases
}

// GetPurchase returns the purchase made by a member in the guild for the given item name.
// If the purchase does not exist, nil is returned.
func GetPurchase(guildID string, memberID string, itemName string, itemType string) *Purchase {
	log.Trace("--> shop.GetPurchase")
	defer log.Trace("<-- shop.GetPurchase")
	purchase, err := readPurchase(guildID, memberID, itemName, itemType)
	if err != nil {
		return nil
	}
	return purchase
}

// PurchaseItem creates a new Purchase with the given guild ID, member ID, and a purchasable
// shop item.
func PurchaseItem(guildID, memberID string, item *ShopItem, renew bool) (*Purchase, error) {

	p := discmsg.GetPrinter(language.AmericanEnglish)

	bankAccount := bank.GetAccount(guildID, memberID)
	err := bankAccount.Withdraw(item.Price)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "member": memberID, "item": item.Name, "error": err}).Error("unable to withdraw cash from the bank account")
		return nil, errors.New(p.Sprintf("insufficient funds to buy the %s `%s` for %d", item.Type, item.Name, item.Price))
	}

	purchase := &Purchase{
		GuildID:     guildID,
		MemberID:    memberID,
		Item:        item,
		Status:      PURCHASED,
		PurchasedOn: time.Now(),
	}
	if item.AutoRenewable {
		purchase.AutoRenew = renew
	}
	if item.Duration != "" {
		duration, _ := disctime.ParseDuration(item.Duration)
		purchase.ExpiresOn = disctime.RoundToNextDay(time.Now().Add(duration))
	}
	err = writePurchase(purchase)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "member": memberID, "item": item.Name, "error": err}).Error("unable to write purchase to the database")
		bankAccount.Deposit(item.Price)
		return nil, fmt.Errorf("unable to write purchase to the database: %w", err)
	}
	log.WithFields(log.Fields{"guild": guildID, "member": memberID, "item": item.Name}).Info("creating new purchase")

	return purchase, nil
}

// Determine if a purchase has expired.
func (p *Purchase) HasExpired() bool {
	log.Trace("--> shop.Purchase.HasExpired")
	defer log.Trace("<-- shop.Purchase.HasExpired")

	if p.IsExpired {
		return true
	}

	oldIsExpired := p.IsExpired
	switch {
	case p.ExpiresOn.IsZero():
		p.IsExpired = false
	case p.ExpiresOn.Before(time.Now()):
		p.IsExpired = true
	default:
		p.IsExpired = false
	}

	if p.IsExpired != oldIsExpired {
		writePurchase(p)
	}

	return p.IsExpired
}

// Return the purchase to the shop.
func (p *Purchase) Return() error {
	log.Trace("--> shop.Purchase.Return")
	defer log.Trace("<-- shop.Purchase.Return")

	bankAccount := bank.GetAccount(p.GuildID, p.MemberID)
	err := bankAccount.Deposit(p.Item.Price)
	if err != nil {
		log.WithFields(log.Fields{"guild": p.GuildID, "member": p.MemberID, "item": p.Item.Name, "error": err}).Error("unable to deposit cash to the bank account")
		return fmt.Errorf("unable to deposit cash to the bank account: %w", err)
	}

	err = deletePurchase(p)
	if err != nil {
		log.WithFields(log.Fields{"guild": p.GuildID, "member": p.MemberID, "item": p.Item.Name, "error": err}).Error("unable to delete purchase from the database")
		return fmt.Errorf("unable to delete purchase from the database: %w", err)
	}

	return nil
}

// Update updates the purchase with the given autoRenew value.
func (p *Purchase) Update(autoRenew bool) error {
	log.Trace("--> shop.Purchase.Update")
	defer log.Trace("<-- shop.Purchase.Update")

	if p.AutoRenew == autoRenew {
		log.WithFields(log.Fields{"guild": p.GuildID, "member": p.MemberID, "item": p.Item.Name}).Info("purchase already has the same autoRenew value")
		return fmt.Errorf("purchase already has the same autoRenew value")
	}

	p.AutoRenew = autoRenew
	err := writePurchase(p)
	if err != nil {
		log.WithFields(log.Fields{"guild": p.GuildID, "member": p.MemberID, "item": p.Item.Name, "error": err}).Error("unable to update purchase in the database")
		return fmt.Errorf("unable to update purchase in the database: %w", err)
	}
	log.WithFields(log.Fields{"guild": p.GuildID, "member": p.MemberID, "item": p.Item.Name}).Info("updating purchase")

	return nil
}

// String returns a string representation of the purchase.
func (p *Purchase) String() string {
	sb := &strings.Builder{}

	sb.WriteString("Purchase{")
	sb.WriteString("GuildID: ")
	sb.WriteString(p.GuildID)
	sb.WriteString(", MemberID: ")
	sb.WriteString(p.MemberID)
	sb.WriteString(", Item: ")
	sb.WriteString(p.Item.String())
	sb.WriteString(", Status: ")
	sb.WriteString(p.Status)
	sb.WriteString(", PurchasedOn: ")
	sb.WriteString(p.PurchasedOn.Format(time.RFC3339))
	if !p.ExpiresOn.IsZero() {
		sb.WriteString(", ExpiresOn: ")
		sb.WriteString(p.ExpiresOn.Format(time.RFC3339))
		sb.WriteString(", AutoRenew: ")
		sb.WriteString(fmt.Sprintf("%v", p.AutoRenew))
		sb.WriteString(", IsExpired: ")
		sb.WriteString(fmt.Sprintf("%v", p.IsExpired))
	}

	return sb.String()
}
