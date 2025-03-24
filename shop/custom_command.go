package shop

import log "github.com/sirupsen/logrus"

const (
	CUSTOM_COMMAND = "custom_command"
)

// CustomCommand represents a custom command item in the shop.
type CustomCommand ShopItem

// GetCustomCommand retrieves a custom command from the shop by its name for a specific guild.
func GetCustomCommand(guildID string, name string) *CustomCommand {
	log.Trace("--> shop.GetCustomCommand")
	defer log.Trace("<-- shop.GetCustomCommand")

	item := GetShopItem(guildID, name, ROLE)
	customCommand := (*CustomCommand)(item)
	return customCommand
}

// NewCustomCommand creates a new custom command item for the shop.
func NewCustomCommand(guildID string, name string, description string, price int, duration string, autoRenewable bool) *CustomCommand {
	log.Trace("--> shop.NewCustomCommand")
	defer log.Trace("<-- shop.NewCustomCommand")

	item := NewShopItem(guildID, name, description, CUSTOM_COMMAND, price, duration, autoRenewable)
	customCommand := CustomCommand(*item)
	return &customCommand
}

// Purchase allows a member to purchase the custom command item from the shop.
func (c *CustomCommand) Purchase(memberID string, renew bool) (*Purchase, error) {
	log.Trace("--> shop.CustomCommand.Purchase")
	defer log.Trace("<-- shop.CustomCommand.Purchase")

	item := ShopItem(*c)
	return item.Purchase(memberID, renew)
}
