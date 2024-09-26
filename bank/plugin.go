package bank

import (
	"fmt"
	"sort"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/dgame/discord"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	PLUGIN_NAME = "bank"
)

var (
	plugin *Plugin
)

// Plugin is the plugin for the banking system used by the bot
type Plugin struct {
	bot *discord.Bot
}

// init creates and registers the plugin for the banking system
func init() {
	plugin = &Plugin{}
	discord.RegisterPlugin(plugin)
}

// Initialize saves the Discord bot to be used by the banking system
func (plugin *Plugin) Initialize(bot *discord.Bot) {
	plugin.bot = bot
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
	return commandHandlers
}

// GetName returns the name of the banking system plugin
func (plugin *Plugin) GetName() string {
	return PLUGIN_NAME
}

// GetHelp returns the member help for the banking system
func (plugin *Plugin) GetHelp() []string {
	help := make([]string, 0, 1)

	for _, command := range memberCommands[0].Options {
		commandDescription := fmt.Sprintf("- **/%s %s**:  %s\n", PLUGIN_NAME, command.Name, command.Description)
		help = append(help, commandDescription)
	}
	sort.Slice(help, func(i, j int) bool {
		return help[i] < help[j]
	})
	title := fmt.Sprintf("**%s**\n", cases.Title(language.English, cases.Compact).String(PLUGIN_NAME))
	help = append([]string{title}, help...)

	return help
}

// GetAdminHelp returns the admin help for the banking system
func (plugin *Plugin) GetAdminHelp() []string {
	help := make([]string, 0, len(adminCommands[0].Options))

	for _, command := range adminCommands[0].Options {
		commandDescription := fmt.Sprintf("- **/%s-admin %s**:  %s\n", PLUGIN_NAME, command.Name, command.Description)
		help = append(help, commandDescription)
	}
	sort.Slice(help, func(i, j int) bool {
		return help[i] < help[j]
	})
	title := fmt.Sprintf("**%s**\n", cases.Title(language.English, cases.Compact).String(PLUGIN_NAME))
	help = append([]string{title}, help...)

	return help
}
