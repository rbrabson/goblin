package shop

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	CUSTOM_COMMAND = "custom_command"
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
func NewCustomCommand(guildID string, name string, description string, price int, duration string, autoRenewable bool) *CustomCommand {
	item := newShopItem(guildID, name, description, CUSTOM_COMMAND, price, duration, autoRenewable)
	command := (*CustomCommand)(item)
	return command
}

// Update updates the command's properties in the shop.
func (cc *CustomCommand) Update(name string, description string, price int, duration string, autoRenewable bool) error {
	item := (*ShopItem)(cc)
	return item.update(name, description, CUSTOM_COMMAND, price, duration, autoRenewable)
}

// Purchase allows a member to purchase the command from the shop.
func (cc *CustomCommand) Purchase(memberID string, renew bool) (*Purchase, error) {
	item := ShopItem(*cc)
	return item.purchase(memberID, renew)
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
func customCommandCreateChecks(s *discordgo.Session, i *discordgo.InteractionCreate, commandName string) error {
	return createChecks(i.GuildID, commandName, CUSTOM_COMMAND)
}

// customCommandPurchaseChecks performs checks to see if a custom command can be purchased.
func customCommandPurchaseChecks(s *discordgo.Session, i *discordgo.InteractionCreate, commandName string) error {
	// Make sure the command is still available in the shop
	shopItem := getShopItem(i.GuildID, commandName, CUSTOM_COMMAND)
	if shopItem == nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "commandName": commandName}).Error("failed to read command from shop")
		return fmt.Errorf("command `%s` not found in the shop", commandName)
	}

	// Make common checks for all purchases
	err := purchaseChecks(i.GuildID, i.Member.User.ID, CUSTOM_COMMAND, commandName)
	if err != nil {
		return err
	}

	return nil
}
