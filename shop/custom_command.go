package shop

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	log "github.com/sirupsen/logrus"
)

const (
	CUSTOM_COMMAND             = "custom_command"
	CUSTOM_COMMAND_NAME        = "Custom Command"
	CUSTOM_COMMAND_DESCRIPTION = "Custom command that may be used on this server"
)

// CustomCommand represents a custom command item in the shop.
type CustomCommand ShopItem

// GetCustomCommand retrieves a custom command from the shop by its name for a specific guild.
func GetCustomCommand(guildID string, name string) *CustomCommand {
	item := getShopItem(guildID, name, CUSTOM_COMMAND)
	command := CustomCommand(*item)
	return &command
}

// NewCustomCommand creates a new command for the shop.
func NewCustomCommand(guildID string, name string, description string, price int) *CustomCommand {
	if description == "" {
		description = fmt.Sprintf("Custom command `%s`", name)
	}
	item := newShopItem(guildID, name, description, CUSTOM_COMMAND, price, "", false)
	command := (*CustomCommand)(item)
	return command
}

// Update updates the command's properties in the shop.
func (cc *CustomCommand) Update(name string, description string, price int, duration string, autoRenewable bool) error {
	item := (*ShopItem)(cc)
	return item.update(name, description, CUSTOM_COMMAND, price, duration, autoRenewable)
}

// Purchase allows a member to purchase the command from the shop.
func (cc *CustomCommand) Purchase(s *discordgo.Session, memberID string) (*Purchase, error) {
	item := ShopItem(*cc)
	purchase, err := item.purchase(memberID, PENDING, false)
	if err != nil {
		log.WithFields(log.Fields{"guildID": cc.GuildID, "commandName": cc.Name, "memberID": memberID}).WithError(err).Error("failed to purchase custom command")
		return nil, err
	}

	config := GetConfig(cc.GuildID)

	// Notify ModMail
	dm := disgomsg.NewDirectMessage(
		disgomsg.WithContent(fmt.Sprintf("Purchase of custom command `%s` has been initiated. Please contact <@%s> to complete the purchase.", cc.Name, config.NotificationID)),
	)
	_, err = dm.Send(s, config.NotificationID)
	if err != nil {
		log.WithFields(log.Fields{"guildID": cc.GuildID, "commandName": cc.Name, "notificationID": config.NotificationID}).WithError(err).Error("failed to send notification message")
		purchase.Return()
		return nil, err
	}

	// Notify the member
	dm = disgomsg.NewDirectMessage(
		disgomsg.WithContent(fmt.Sprintf("Purchase of custom command `%s` has been initiated. Please contact <@%s> to complete the purchase.", cc.Name, config.NotificationID)),
	)
	_, err = dm.Send(s, memberID)
	if err != nil {
		log.WithFields(log.Fields{"guildID": cc.GuildID, "commandName": cc.Name, "memberID": memberID}).WithError(err).Error("failed to send direct message")
		purchase.Return()
		return nil, err
	}

	return purchase, nil
}

// AddToShop adds the command to the shop. If the command already exists, an error is returned.
func (cc *CustomCommand) AddToShop(s *Shop) error {
	item := (*ShopItem)(cc)
	return item.addToShop(s)
}

// RemoveFromShop removes the command from the shop. If the command does not exist, an error is returned.
func (cc *CustomCommand) RemoveFromShop(s *Shop) error {
	item := (*ShopItem)(cc)
	return item.removeFromShop(s)
}

// customCommandCreateChecks performs checkst to see if a custom command can be added to the shop.
func customCommandCreateChecks(guildID string, commandName string) error {
	return createChecks(guildID, commandName, CUSTOM_COMMAND)
}

// customCommandPurchaseChecks performs checks to see if a role can be purchased.
func customCommandPurchaseChecks(s *discordgo.Session, i *discordgo.InteractionCreate, commandName string) error {
	// Make sure the role is still available in the shop
	shopItem := getShopItem(i.GuildID, commandName, ROLE)
	if shopItem == nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "commandName": commandName}).Error("failed to read custom command from shop")
		return fmt.Errorf("custom command `%s` not found in the shop", commandName)
	}

	// Make common checks for all purchases
	err := purchaseChecks(i.GuildID, i.Member.User.ID, CUSTOM_COMMAND, commandName)
	if err != nil {
		return err
	}

	return nil
}
