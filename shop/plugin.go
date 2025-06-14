package shop

import (
	"fmt"
	"slices"

	"github.com/bwmarrin/discordgo"
	page "github.com/rbrabson/disgopage"
	"github.com/rbrabson/goblin/database/mongo"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	PluginName       = "shop"
	PurchasesPerPage = 5
)

var (
	plugin *Plugin
	bot    *discord.Bot
	db     *mongo.MongoDB
	status = discord.RUNNING
)

// Plugin is the plugin for the banking system used by the bot
type Plugin struct{}

// Ensure the plugin implements the Plugin interface
var _ discord.Plugin = (*Plugin)(nil)

// Start creates and registers the plugin for the banking system
func Start() {
	plugin = &Plugin{}
	discord.RegisterPlugin(plugin)
}

// Stop stops the heist game. This is called when the bot is shutting down.
func (plugin *Plugin) Stop() {
	paginator.Close()
	status = discord.STOPPED
}

// Status returns the status of the heist game.	This is used to determine
// if the plugin is running or not.
func (plugin *Plugin) Status() discord.PluginStatus {
	return status
}

// Initialize saves the Discord bot to be used by the banking system
func (plugin *Plugin) Initialize(b *discord.Bot, d *mongo.MongoDB) {
	bot = b
	db = d
	registerAllShoopItemComponentHandlers()
	paginator = page.NewPaginator(
		page.WithDiscordConfig(
			page.DiscordConfig{
				Session:                bot.Session,
				AddComponentHandler:    bot.AddComponentHandler,
				RemoveComponentHandler: bot.RemoveComponentHandler,
			},
		),
		page.WithItemsPerPage(PurchasesPerPage),
	)
	go checkForExpiredPurchases()
}

// GetCommands returns the commands for the banking system
func (plugin *Plugin) GetCommands() []*discordgo.ApplicationCommand {
	commands := make([]*discordgo.ApplicationCommand, 0, len(adminCommands)+len(memberCommands))
	commands = append(commands, adminCommands...)
	commands = append(commands, memberCommands...)
	return commands
}

// GetCommandHandlers returns the command handlers for the banking system
func (plugin *Plugin) GetCommandHandlers() map[string]func(*discordgo.Session, *discordgo.InteractionCreate) {
	return commandHandlers
}

// GetComponentHandlers returns the component handlers for the banking system
func (plugin *Plugin) GetComponentHandlers() map[string]func(*discordgo.Session, *discordgo.InteractionCreate) {
	return nil
}

// GetName returns the name of the banking system plugin
func (plugin *Plugin) GetName() string {
	return PluginName
}

// GetHelp returns the member help for the banking system
func (plugin *Plugin) GetHelp() []string {
	help := make([]string, 0, 1)

	commandPrefix := memberCommands[0].Name
	for _, command := range memberCommands[0].Options {
		commandDescription := fmt.Sprintf("- `/%s %s`: %s\n", commandPrefix, command.Name, command.Description)
		help = append(help, commandDescription)
	}
	slices.Sort(help)
	title := fmt.Sprintf("## %s\n", cases.Title(language.AmericanEnglish, cases.Compact).String(PluginName))
	help = append([]string{title}, help...)

	return help
}

// GetAdminHelp returns the admin help for the banking system
func (plugin *Plugin) GetAdminHelp() []string {
	help := make([]string, 0, len(adminCommands[0].Options))

	commandPrefix := adminCommands[0].Name
	for _, command := range adminCommands[0].Options {
		commandDescription := fmt.Sprintf("- `/%s %s`: %s\n", commandPrefix, command.Name, command.Description)
		help = append(help, commandDescription)
	}
	slices.Sort(help)
	title := fmt.Sprintf("## %s\n", cases.Title(language.AmericanEnglish, cases.Compact).String(PluginName))
	help = append([]string{title}, help...)

	return help
}

// registerAllShoopItemComponentHandlers adds the component handlers for all
// shop items that may be purchased.
func registerAllShoopItemComponentHandlers() {
	for _, g := range guild.GetAllGuilds() {
		shop := GetShop(g.GuildID)
		for _, item := range shop.Items {
			registerShopItemComponentHandlers(item)
		}
	}
}

// registerShopItemComponentHandlers registers the component handlers for
// the shop item.
func registerShopItemComponentHandlers(shopItem *ShopItem) {
	// Register the component handlers for the item
	customID := shopItem.Type + ":" + shopItem.Name
	bot.AddComponentHandler("shop:"+customID, initiatePurchase)
	bot.AddComponentHandler("shop:buy:"+customID, completePurchase)
}
