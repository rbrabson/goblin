package bank

import (
	"fmt"
	"sort"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/dgame/discord"
)

const (
	PLUGIN_NAME = "bank"
)

var (
	plugin *BankPlugin
)

// BankPlugin is the plugin for the banking system used by the bot
type BankPlugin struct {
	bot *discord.Bot
}

// init creates and registers the plugin for the banking system
func init() {
	plugin = &BankPlugin{}
	discord.RegisterPlugin(plugin)
}

// Initialize saves the Discord bot to be used by the banking system
func (plugin *BankPlugin) Initialize(bot *discord.Bot) {
	plugin.bot = bot
}

// GetCommands returns the commands for the banking system
func (plugin *BankPlugin) GetCommands() []*discordgo.ApplicationCommand {
	commands := make([]*discordgo.ApplicationCommand, 0, len(adminCommands)+len(memberCommands))
	commands = append(commands, adminCommands...)
	commands = append(commands, memberCommands...)
	return commands
}

// GetCommandHandlers returns the command handlers for the banking system
func (plugin *BankPlugin) GetCommandHandlers() map[string]func(*discordgo.Session, *discordgo.InteractionCreate) {
	return commandHandlers
}

// GetComponentHandlers returns the component handlers for the banking system
func (plugin *BankPlugin) GetComponentHandlers() map[string]func(*discordgo.Session, *discordgo.InteractionCreate) {
	return commandHandlers
}

// GetName returns the name of the banking system plugin
func (plugin *BankPlugin) GetName() string {
	return PLUGIN_NAME
}

// GetHelp returns the member help for the banking system
func (plugin *BankPlugin) GetHelp() []string {
	help := make([]string, 0, 1)

	for _, command := range memberCommands[0].Options {
		commandDescription := fmt.Sprintf("- **/bank %s**:  %s\n", command.Name, command.Description)
		help = append(help, commandDescription)
	}
	sort.Slice(help, func(i, j int) bool {
		return help[i] < help[j]
	})
	help = append([]string{"**Bank**\n"}, help...)

	return help
}

// GetAdminHelp returns the admin help for the banking system
func (plugin *BankPlugin) GetAdminHelp() []string {
	help := make([]string, 0, len(adminCommands[0].Options))

	for _, command := range adminCommands[0].Options {
		commandDescription := fmt.Sprintf("- **/bank-admin %s**:  %s\n", command.Name, command.Description)
		help = append(help, commandDescription)
	}
	sort.Slice(help, func(i, j int) bool {
		return help[i] < help[j]
	})
	help = append([]string{"**Bank**\n"}, help...)

	return help
}
