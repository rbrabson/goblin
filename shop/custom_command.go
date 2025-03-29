package shop

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
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
func NewCustomCommand(guildID string, price int) *CustomCommand {
	item := newShopItem(guildID, "Custom Command", "Custom command that may be used on this server", CUSTOM_COMMAND, price, "", false)
	command := (*CustomCommand)(item)
	return command
}

// Update updates the command's properties in the shop.
func (cc *CustomCommand) Update(name string, description string, price int, duration string, autoRenewable bool) error {
	item := (*ShopItem)(cc)
	return item.update(name, description, CUSTOM_COMMAND, price, duration, autoRenewable)
}

// Purchase allows a member to purchase the command from the shop.
func (cc *CustomCommand) Purchase(s *discordgo.Session, memberID string, renew bool) (*Purchase, error) {
	item := ShopItem(*cc)
	purchase, err := item.purchase(memberID, PENDING, renew)
	if err != nil {
		log.WithFields(log.Fields{"guildID": cc.GuildID, "commandName": cc.Name, "memberID": memberID, "renew": renew}).WithError(err).Error("failed to purchase command")
		return nil, err
	}

	dm := disgomsg.DirectMessage{
		Content: "Test Message",
	}
	err = dm.Send(s, memberID)
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
func customCommandCreateChecks(s *discordgo.Session, i *discordgo.InteractionCreate, commandName string) error {
	return createChecks(i.GuildID, commandName, CUSTOM_COMMAND)
}
