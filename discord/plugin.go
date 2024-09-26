package discord

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

var (
	plugins = make([]Plugin, 0, 1)
	mutex   sync.Mutex
)

// Plugin defines the game that is registered to run on the system
type Plugin interface {
	Initialize(bot *Bot)
	GetCommandHandlers() map[string]func(*discordgo.Session, *discordgo.InteractionCreate)
	GetCommands() []*discordgo.ApplicationCommand
	GetComponentHandlers() map[string]func(*discordgo.Session, *discordgo.InteractionCreate)
	GetHelp() []string
	GetName() string
	GetAdminHelp() []string
}

// ListPlugin returns the list of plugins that have been registered for use within the bot
func ListPlugin() []Plugin {
	return plugins
}

// RegisterPlugin registers the plugin to be used within the bot
func RegisterPlugin(plugin Plugin) {
	mutex.Lock()
	defer mutex.Unlock()

	plugins = append(plugins, plugin)
}
