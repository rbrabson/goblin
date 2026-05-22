package discord

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/database/mongo"
)

var (
	plugins = make([]Plugin, 0, 10)
	mutex   sync.Mutex
)

type PluginStatus int

const (
	PluginRunning PluginStatus = iota
	PluginStopping
	PluginStopped
)

// Plugin defines the game that is registered to run on the system
type Plugin interface {
	Initialize(bot *Bot, db *mongo.MongoDB)
	GetCommandHandlers() map[string]func(*discordgo.Session, *discordgo.InteractionCreate)
	GetCommands() []*discordgo.ApplicationCommand
	GetComponentHandlers() map[string]func(*discordgo.Session, *discordgo.InteractionCreate)
	GetHelp() []string
	GetName() string
	GetAdminHelp() []string
	Stop()
	Status() PluginStatus
}

// ListPlugin returns the list of plugins that have been registered for use within the bot
func ListPlugin() []Plugin {
	mutex.Lock()
	defer mutex.Unlock()

	return append([]Plugin(nil), plugins...)
}

// RegisterPlugin registers the plugin to be used within the bot
func RegisterPlugin(plugin Plugin) {
	mutex.Lock()
	defer mutex.Unlock()

	plugins = append(plugins, plugin)
}

// String gets the string representation of the plugin status.
func (s PluginStatus) String() string {
	switch s {
	case PluginRunning:
		return "Running"
	case PluginStopping:
		return "Stopping"
	case PluginStopped:
		return "Stopped"
	default:
		return "Unknown"
	}
}
