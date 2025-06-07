package payday

import (
	"fmt"
	"slices"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/database/mongo"
	"github.com/rbrabson/goblin/discord"
)

const (
	PLUGIN_NAME = "payday"
)

var (
	plugin *Plugin
	db     *mongo.MongoDB
	status = discord.RUNNING
)

// Plugin is the plugin for the payday system used by the bot
type Plugin struct{}

// Ensure the plugin implements the Plugin interface
var _ discord.Plugin = (*Plugin)(nil)

// Start creates and registers the plugin for the payday system
func Start() {
	plugin = &Plugin{}
	discord.RegisterPlugin(plugin)
}

// Stop stops the heist game. This is called when the bot is shutting down.
func (plugin *Plugin) Stop() {
	status = discord.STOPPED
}

// Status returns the status of the heist game.	This is used to determine
// if the plugin is running or not.
func (plugin *Plugin) Status() discord.PluginStatus {
	return status
}

// GetMemberHelp returns help information about the heist bot commands
func GetMemberHelp() []string {
	help := make([]string, 0, 1)

	for _, command := range memberCommands {
		commandDescription := fmt.Sprintf("- `/%s`: %s\n", command.Name, command.Description)
		help = append(help, commandDescription)
	}
	slices.Sort(help)
	title := fmt.Sprintf("## %s\n", cases.Title(language.AmericanEnglish, cases.Compact).String(PLUGIN_NAME))
	help = append([]string{title}, help...)

	return help
}

// GetAdminHelp returns help information about the heist bot commands
func GetAdminHelp() []string {
	return nil
}

// Initialize saves the Discord bot to be used by the banking system
func (plugin *Plugin) Initialize(b *discord.Bot, d *mongo.MongoDB) {
	db = d
}

// GetCommands returns the commands for the banking system
func (plugin *Plugin) GetCommands() []*discordgo.ApplicationCommand {
	commands := make([]*discordgo.ApplicationCommand, 0, len(memberCommands))
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
	return PLUGIN_NAME
}

// GetHelp returns the member help for the banking system
func (plugin *Plugin) GetHelp() []string {
	help := make([]string, 0, 1)

	for _, command := range memberCommands {
		commandDescription := fmt.Sprintf("- `/%s`: %s\n", command.Name, command.Description)
		help = append(help, commandDescription)
	}
	slices.Sort(help)
	title := fmt.Sprintf("## %s\n", cases.Title(language.AmericanEnglish, cases.Compact).String(PLUGIN_NAME))
	help = append([]string{title}, help...)

	return help
}

// GetAdminHelp returns the admin help for the banking system
func (plugin *Plugin) GetAdminHelp() []string {
	return nil
}
