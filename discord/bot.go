package discord

import (
	"log/slog"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	"github.com/rbrabson/goblin/database/mongo"
	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/internal/logger"
)

const (
	botIntents = discordgo.IntentGuilds |
		discordgo.IntentGuildMessages |
		discordgo.IntentDirectMessages |
		discordgo.IntentGuildEmojis
)

var (
	Version  string
	Revision string
	BotName  = "Goblin"
	db       *mongo.MongoDB
)

var (
	sslog = logger.GetLogger()
)

var (
	componentHandlers       = make(map[string]func(*discordgo.Session, *discordgo.InteractionCreate))
	commandHandlers         = make(map[string]func(*discordgo.Session, *discordgo.InteractionCreate))
	commands                = make([]*discordgo.ApplicationCommand, 0, 2)
	customComponentHandlers = make(map[string]func(*discordgo.Session, *discordgo.InteractionCreate))
)

// Bot is a Discord bot which is capable of running multiple services, each of which
// implement various commands.
type Bot struct {
	Session *discordgo.Session
	DB      mongo.MongoDB
	appID   string
	guildID string
	timer   chan int
}

// NewBot creates a nbew Discord bot that can run Discord commands.
func NewBot(botName string, version string, revision string) *Bot {
	// Get the bot version and revision
	BotName = botName
	Version = version
	Revision = revision

	// Use environment variables to pass in sensitive data used to identify the bot
	appID := os.Getenv("DISCORD_APP_ID")
	token := os.Getenv("DISCORD_BOT_TOKEN")
	guildID := os.Getenv("DISCORD_GUILD_ID")

	s, err := discordgo.New("Bot " + token)
	if err != nil {
		sslog.Error("failed to create the bot",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	bot := &Bot{
		Session: s,
		timer:   make(chan int),
		appID:   appID,
		guildID: guildID,
	}
	bot.Session.Identify.Intents = botIntents

	bot.Session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		sslog.Info("Bot is up!",
			slog.String("bot", BotName),
			slog.String("version", Version),
			slog.String("revision", Revision),
		)
	})

	db = mongo.NewDatabase()
	guild.SetDB(db)
	for _, plugin := range ListPlugin() {
		plugin.Initialize(bot, db)
		sslog.Info("initialized plugin",
			slog.String("plugin", plugin.GetName()),
		)
	}

	// Add commands and handlers for the bot itself
	commands = append(commands, helpCommands...)
	for key, value := range helpCommandHandler {
		commandHandlers[key] = value
	}
	commands = append(commands, serverCommands...)
	for key, value := range serverCommandHandler {
		commandHandlers[key] = value
	}

	// Add commands and handlers for each plugin
	for _, plugin := range ListPlugin() {
		commands = append(commands, plugin.GetCommands()...)
		for key, handler := range plugin.GetCommandHandlers() {
			commandHandlers[key] = handler
		}
		for key, handler := range plugin.GetComponentHandlers() {
			componentHandlers[key] = handler
		}
	}

	// Register a function to add the command or component handler for each plugin
	sslog.Debug("add bot handlers")
	bot.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			} else {
				sslog.Warn("unhandled command",
					slog.String("command", i.ApplicationCommandData().Name),
				)
				resp := disgomsg.NewResponse(
					disgomsg.WithContent("Unknown command. Use `/help` to see a list of available commands."),
				)
				resp.SendEphemeral(s, i.Interaction)
			}
		case discordgo.InteractionMessageComponent:
			if h, ok := componentHandlers[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			} else {
				if h, ok := customComponentHandlers[i.MessageComponentData().CustomID]; ok {
					h(s, i)
				} else {
					sslog.Warn("unhandled component",
						slog.String("component", i.MessageComponentData().CustomID),
					)
					resp := disgomsg.NewResponse(
						disgomsg.WithContent("Unknown component. Please try again."),
					)
					resp.SendEphemeral(s, i.Interaction)
				}
			}
		}
	})
	sslog.Debug("bot handlers added")

	deleteSlashCommands := GetenvBool("DISCORD_DELETE_SLASH_COMMANDS")
	if deleteSlashCommands {
		bot.DeleteCommands()
	}
	bot.LoadCommands(commands)

	return bot
}

// DeleteCommands deletes the current set of slash commands. This can be useful when developing
// a new bot and the set of loaded slash commands changes.
func (bot *Bot) DeleteCommands() {
	// Delete all bot commands indivdually
	// commands, err := bot.Session.ApplicationCommands(bot.appID, bot.guildID)
	// if err != nil {
	// 	log.WithFields(log.Fields{"appID": bot.appID, "guildID": bot.guildID, "error": err}).Fatal("failed to get bot commands")
	// }
	// for _, command := range commands {
	// 	log.WithFields(log.Fields{"name": command.Name, "description": command.Description}).Debug("deleting command")
	// 	err := bot.Session.ApplicationCommandDelete(bot.appID, bot.guildID, command.ID)
	// 	if err != nil {
	// 		log.WithFields(log.Fields{"name": command.Name, "description": command.Description, "error": err}).Error("failed to delete command")
	// 	}
	// }

	sslog.Debug("deleting old bot commands")
	_, err := bot.Session.ApplicationCommandBulkOverwrite(bot.appID, bot.guildID, nil)
	if err != nil {
		sslog.Error("failed to delete old commands", slog.String("error", err.Error()))
		os.Exit(1)
	}
	sslog.Debug("old bot commands deleted")
}

// LoadCommands register all the commands. This implicitly will call the function added above that
// adds the command and component handlers for each plugin.
func (bot *Bot) LoadCommands(commands []*discordgo.ApplicationCommand) {
	sslog.Debug("load new bot commands",
		slog.String("appID", bot.appID),
		slog.String("guildID", bot.guildID),
	)
	_, err := bot.Session.ApplicationCommandBulkOverwrite(bot.appID, bot.guildID, commands)
	if err != nil {
		for _, command := range commands {
			sslog.Error("failed to load command",
				slog.String("name", command.Name),
				slog.String("description", command.Description),
			)
		}
		sslog.Error("failed to load bot commands",
			slog.String("appID", bot.appID),
			slog.String("guildID", bot.guildID),
			slog.String("error", err.Error()),
			"commands", commands,
		)
		os.Exit(1)

	}
	sslog.Info("new bot commands loaded")
}

// AddComponentHandler adds a component handler for the bot. This is used to handle
// components that are not explicitly defined in the bot.
func (bot *Bot) AddComponentHandler(key string, handler func(*discordgo.Session, *discordgo.InteractionCreate)) {
	customComponentHandlers[key] = handler
}

// removeComponentHandler removes a component handler for the bot. This is used to remove
// components that are not explicitly defined in the bot.
func (bot *Bot) RemoveComponentHandler(key string) {
	delete(customComponentHandlers, key)
}
