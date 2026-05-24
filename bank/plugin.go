package bank

import (
	"fmt"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/database/mongo"
	"github.com/rbrabson/goblin/discord"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	PluginName = "bank"
)

var (
	plugin *Plugin
	db     *mongo.MongoDB
	status = discord.PluginRunning
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

// Initialize saves the Discord bot to be used by the banking system
func (plugin *Plugin) Initialize(b *discord.Bot, d *mongo.MongoDB) {
	db = d
}

// Stop stops the banking system. This is called when the bot is shutting down.
func (plugin *Plugin) Stop() {
	status = discord.PluginStopped
}

// Status returns the status of the banking system.	This is used to determine
// if the plugin is running or not.
func (plugin *Plugin) Status() discord.PluginStatus {
	return status
}

// SetDB sets the database for testing purposes
func SetDB(d *mongo.MongoDB) {
	db = d
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
	return getHelp(memberCommands)
}

// GetAdminHelp returns the admin help for the banking system
func (plugin *Plugin) GetAdminHelp() []string {
	return getHelp(adminCommands)
}

// getHelp returns the help for the banking system based on the provided commands. It formats the help as a list of
// strings, with each string representing a command and its description. The help is sorted alphabetically by command
// name and includes a title for the plugin.
func getHelp(commands []*discordgo.ApplicationCommand) []string {
	var help = make([]string, 0, len(commands[0].Options))

	commandPrefix := commands[0].Name
	for _, command := range commands[0].Options {
		commandDescription := fmt.Sprintf("- `/%s %s`: %s\n", commandPrefix, command.Name, command.Description)
		help = append(help, commandDescription)
	}
	slices.Sort(help)
	title := fmt.Sprintf("## %s\n", cases.Title(language.AmericanEnglish, cases.Compact).String(PluginName))
	help = append([]string{title}, help...)

	return help
}
