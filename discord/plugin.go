package discord

import (
	"log/slog"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
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

// getBotStatus returns the status of the bot.
func getBotStatus() PluginStatus {
	botStatus := PluginRunning
	for _, plugin := range plugins {
		switch plugin.Status() {
		case PluginStopping:
			botStatus = PluginStopping
		case PluginStopped:
			if botStatus == PluginRunning {
				botStatus = PluginStopped
			}
		default:
			// NO-OP
		}
	}
	return botStatus
}

// IsShuttingDown returns true if the bot is shutting down. It also sends an ephemeral message to the user if the bot
// is shutting down.
func IsShuttingDown(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	status := getBotStatus()
	shuttingDown := status != PluginRunning

	if shuttingDown {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The system is shutting down."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("error sending response",
				slog.String("guildID", i.GuildID),
				slog.String("error", err.Error()),
			)
		}
	}

	return shuttingDown
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
