package blackjack

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
	PluginName = "blackjack"
)

var (
	plugin *Plugin
	db     *mongo.MongoDB
	bot    *discord.Bot
	status = discord.RUNNING
)

// Plugin is the plugin for the slots system used by the bot
type Plugin struct{}

// Ensure the plugin implements the Plugin interface
var _ discord.Plugin = (*Plugin)(nil)

// Start creates and registers the plugin for the slots system
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

// Initialize saves the Discord bot to be used by the slots system
func (plugin *Plugin) Initialize(b *discord.Bot, d *mongo.MongoDB) {
	bot = b
	db = d
}

// SetDB sets the database to be used by the slots system. This is used for testing.
func SetDB(d *mongo.MongoDB) {
	db = d
}

// GetCommands returns the commands for the slots system
func (plugin *Plugin) GetCommands() []*discordgo.ApplicationCommand {
	commands := make([]*discordgo.ApplicationCommand, 0, len(memberCommands))
	commands = append(commands, memberCommands...)
	return commands
}

// GetCommandHandlers returns the command handlers for the slots system
func (plugin *Plugin) GetCommandHandlers() map[string]func(*discordgo.Session, *discordgo.InteractionCreate) {
	return commandHandlers
}

// GetComponentHandlers returns the component handlers for the slots system
func (plugin *Plugin) GetComponentHandlers() map[string]func(*discordgo.Session, *discordgo.InteractionCreate) {
	return componentHandlers
}

// GetName returns the name of the slots system plugin
func (plugin *Plugin) GetName() string {
	return PluginName
}

// GetHelp returns the member help for the slots system
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

// GetAdminHelp returns the admin help for the slots system
func (plugin *Plugin) GetAdminHelp() []string {
	return nil
}
