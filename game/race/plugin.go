package race

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
	PluginName = "race"
)

var (
	plugin *Plugin
	bot    *discord.Bot
	db     *mongo.MongoDB
	status = discord.RUNNING
)

// Plugin is the plugin for the heist game
type Plugin struct{}

// Ensure the plugin implements the Plugin interface
var _ discord.Plugin = (*Plugin)(nil)

// Start creates and registers the plugin for the heist game
func Start() {
	plugin = &Plugin{}
	discord.RegisterPlugin(plugin)
}

// Stop stops the heist game. This is called when the bot is shutting down.
func (plugin *Plugin) Stop() {
	status = discord.STOPPING
}

// Status returns the status of the heist game.	This is used to determine
// if the plugin is running or not.
func (plugin *Plugin) Status() discord.PluginStatus {
	if status == discord.STOPPING {
		raceLock.Lock()
		if len(currentRaces) == 0 {
			status = discord.STOPPED
		}
		raceLock.Unlock()
	}

	return status
}

// Initialize saves the Discord bot to be used by the banking system
func (plugin *Plugin) Initialize(b *discord.Bot, d *mongo.MongoDB) {
	bot = b
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
	return componentHandlers
}

// GetName returns the name of the banking system plugin
func (plugin *Plugin) GetName() string {
	return PluginName
}

// GetHelp returns the member help for the banking system
func (plugin *Plugin) GetHelp() []string {
	help := make([]string, 0, 1)

	for _, command := range memberCommands[0].Options {
		commandDescription := fmt.Sprintf("- `/%s %s`: %s\n", PluginName, command.Name, command.Description)
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
