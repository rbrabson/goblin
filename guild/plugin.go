package guild

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/database/mongo"
	"github.com/rbrabson/goblin/discord"
)

const (
	PLUGIN_NAME = "guild"
)

var (
	plugin *Plugin
	db     *mongo.MongoDB
)

// Plugin is the plugin for the guild and it's members
type Plugin struct{}

// Start creates and registers the plugin for the guild and it's members
func Start() {
	plugin = &Plugin{}
	discord.RegisterPlugin(plugin)
}

// Initialize saves the Discord bot to be used by the guild system
func (plugin *Plugin) Initialize(b *discord.Bot, d *mongo.MongoDB) {
	db = d
}

// SetDB sets the database for testing purposes
func SetDB(d *mongo.MongoDB) {
	db = d
}

// GetCommands returns the commands for the guild system
func (plugin *Plugin) GetCommands() []*discordgo.ApplicationCommand {
	return nil
}

// GetCommandHandlers returns the command handlers for the guild system
func (plugin *Plugin) GetCommandHandlers() map[string]func(*discordgo.Session, *discordgo.InteractionCreate) {
	return nil
}

// GetComponentHandlers returns the component handlers for the guild system
func (plugin *Plugin) GetComponentHandlers() map[string]func(*discordgo.Session, *discordgo.InteractionCreate) {
	return nil
}

// GetName returns the name of the guild system plugin
func (plugin *Plugin) GetName() string {
	return PLUGIN_NAME
}

// GetHelp returns the member help for the guild system
func (plugin *Plugin) GetHelp() []string {
	return nil
}

// GetAdminHelp returns the admin help for the guild system
func (plugin *Plugin) GetAdminHelp() []string {
	return nil
}
