package shop

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/internal/disctime"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
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
	AutoRenew   bool               `json:"auto_renew" bson:"auto_renew"`
	IsExpired   bool               `json:"is_expired" bson:"is_expired"`
}

// GetAllRoles returns all the purchases made by a member in the guild.
func GetAllPurchases(guildID string, memberID string) []*Purchase {
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
	purchase, err := readPurchase(guildID, memberID, itemName, itemType)
	if err != nil {
		return nil
	}
	return purchase
}

// PurchaseItem creates a new Purchase with the given guild ID, member ID, and a purchasable
// shop item.
func PurchaseItem(guildID, memberID string, item *ShopItem, renew bool) (*Purchase, error) {
	p := message.NewPrinter(language.AmericanEnglish)

	bankAccount := bank.GetAccount(guildID, memberID)
	err := bankAccount.WithdrawFromCurrentOnly(item.Price)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "member": memberID, "item": item.Name, "error": err}).Warn("unable to withdraw cash from the bank account")
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
	config := GetConfig(guildID)
	if config.ModChannelID != "" {
		guildMember := guild.GetMember(guildID, memberID)
		msg := disgomsg.Message{
			Content: p.Sprintf("`%s` (id=%s) purchased %s `%s` for %d", guildMember.Name, memberID, item.Type, item.Name, item.Price),
		}
		msg.Send(bot.Session, config.ModChannelID)
	}

	return purchase, nil
}

// Determine if a purchase has expired. This marks the purchase as expired and undoes the effects of the purchase
// if it has expired.
func (p *Purchase) HasExpired() bool {
	if p.IsExpired {
		log.WithFields(log.Fields{"guild": p.GuildID, "member": p.MemberID, "item": p.Item.Name}).Trace("purchase has already been marked as expired")
		return true
	}

	oldIsExpired := p.IsExpired
	switch {
	case p.ExpiresOn.IsZero():
		log.WithFields(log.Fields{"guild": p.GuildID, "member": p.MemberID, "item": p.Item.Name}).Trace("purchase has a zero expiration timer")
		return false
	case p.ExpiresOn.Before(time.Now().UTC()):
		switch p.Item.Type {
		case ROLE:
			// Assign the role to the user. If the role can't be assigned, then undo the purchase of the role.
			err := guild.UnAssignRole(bot.Session, p.GuildID, p.MemberID, p.Item.Name)
			if err != nil {
				log.WithFields(log.Fields{"guildID": p.GuildID, "roleName": p.Item.Name, "memberID": p.MemberID, "error": err}).Error("failed to unassign role")
				return false
			}
		default:
			log.WithFields(log.Fields{"guild": p.GuildID, "member": p.MemberID, "item": p.Item.Name}).Info("unknown purchase has expired")
		}

		p.IsExpired = true
	default:
		return false
	}

	if p.IsExpired != oldIsExpired {
		writePurchase(p)

		g, _ := bot.Session.Guild(p.GuildID)
		var msg string
		if g != nil && g.Name != "" {
			msg = fmt.Sprintf("Your purchase of %s `%s` on `%s` has expired", p.Item.Type, p.Item.Name, g.Name)
		} else {
			msg = fmt.Sprintf("Your purchase of %s `%s` has expired", p.Item.Type, p.Item.Name)
		}
		SendMessageToUser(bot.Session, p.MemberID, msg)

		config := GetConfig(p.GuildID)
		if config.ModChannelID != "" {
			guildMember := guild.GetMember(p.GuildID, p.MemberID)
			printer := message.NewPrinter(language.AmericanEnglish)
			msg := disgomsg.Message{
				Content: printer.Sprintf("`%s` (id=%s) had their purchase of %s `%s` expire", guildMember.Name, p.MemberID, p.Item.Type, p.Item.Name),
			}
			msg.Send(bot.Session, config.ModChannelID)
			log.WithFields(log.Fields{"guild": p.GuildID, "member": p.MemberID, "item": p.Item.Name}).Info("purchase has expired")
		} else {
			log.WithFields(log.Fields{"guild": p.GuildID, "member": p.MemberID, "item": p.Item.Name}).Info("no mod channel configured to notify of expired purchase")
		}
	} else {
		log.WithFields(log.Fields{"guild": p.GuildID, "member": p.MemberID, "item": p.Item.Name}).Trace("purchase expiration has not changed")
	}

	return p.IsExpired
}

// Return the purchase to the shop.
func (p *Purchase) Return() error {
	bankAccount := bank.GetAccount(p.GuildID, p.MemberID)
	err := bankAccount.DepositToCurrentOnly(p.Item.Price)
	if err != nil {
		log.WithFields(log.Fields{"guild": p.GuildID, "member": p.MemberID, "item": p.Item.Name, "error": err}).Error("unable to deposit cash to the bank account")
		return fmt.Errorf("unable to deposit cash to the bank account: %w", err)
	}

	err = deletePurchase(p)
	if err != nil {
		log.WithFields(log.Fields{"guild": p.GuildID, "member": p.MemberID, "item": p.Item.Name, "error": err}).Error("unable to delete purchase from the database")
		return fmt.Errorf("unable to delete purchase from the database: %w", err)
	}

	config := GetConfig(p.GuildID)
	if config.ModChannelID != "" {
		guildMember := guild.GetMember(p.GuildID, p.MemberID)
		printer := message.NewPrinter(language.AmericanEnglish)
		msg := disgomsg.Message{
			Content: printer.Sprintf("`%s` (id=%s) has returned the purchase of %s `%s`", guildMember.Name, p.MemberID, p.Item.Type, p.Item.Name),
		}
		msg.Send(bot.Session, config.ModChannelID)
	}

	return nil
}

// Update updates the purchase with the given autoRenew value.
func (p *Purchase) Update(autoRenew bool) error {
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

// checkForExpiredPurchases checks once a day to see if any purchases that may be expired have expired.
func checkForExpiredPurchases() {
	for {
		filter := bson.D{
			{Key: "is_expired", Value: false},
			{Key: "$and", Value: bson.A{
				bson.D{{Key: "expires_on", Value: bson.D{{Key: "$ne", Value: time.Time{}}}}},
				bson.D{{Key: "expires_on", Value: bson.D{{Key: "$lte", Value: time.Now().UTC()}}}},
			}},
		}
		log.WithFields(log.Fields{"filter": filter}).Trace("checking for expired purchases")
		purchases, _ := readAllPurchases(filter)
		log.WithFields(log.Fields{"count": len(purchases)}).Debug("checking for expired purchases")
		for _, purchase := range purchases {
			purchase.HasExpired()
		}

		// Wait until tomorrow to check again
		year, month, day := time.Now().UTC().Date()
		tomorrow := time.Date(year, month, day+1, 0, 0, 0, 0, time.UTC)
		log.WithFields(log.Fields{"tomorrow": tomorrow}).Trace("waiting until tomorrow to check for expired purchases")
		time.Sleep(time.Until(tomorrow))
	}
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
