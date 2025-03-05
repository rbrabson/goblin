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
	PLUGIN_NAME = "bank"
)

var (
	plugin *Plugin
	db     *mongo.MongoDB
)

// Plugin is the plugin for the banking system used by the bot
type Plugin struct{}

// creates and registers the plugin for the banking system
func Start() {
	plugin = &Plugin{}
	discord.RegisterPlugin(plugin)
}

// Initialize saves the Discord bot to be used by the banking system
func (plugin *Plugin) Initialize(b *discord.Bot, d *mongo.MongoDB) {
	db = d
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
	return PLUGIN_NAME
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
	title := fmt.Sprintf("## %s\n", cases.Title(language.AmericanEnglish, cases.Compact).String(PLUGIN_NAME))
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
	title := fmt.Sprintf("## %s\n", cases.Title(language.AmericanEnglish, cases.Compact).String(PLUGIN_NAME))
	help = append([]string{title}, help...)

	return help
}
