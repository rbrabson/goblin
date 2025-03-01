package discord

import (
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/database/mongo"
	"github.com/rbrabson/goblin/guild"
	log "github.com/sirupsen/logrus"
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
	log.Trace("--> discord.NewBot")
	defer log.Trace("<-- discord.NewBot")

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
		log.WithField("error", err).Fatal("failed to create the bot")
	}

	bot := &Bot{
		Session: s,
		timer:   make(chan int),
		appID:   appID,
		guildID: guildID,
	}
	bot.Session.Identify.Intents = botIntents

	bot.Session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info("Bot is up!")
	})

	db = mongo.NewDatabase()
	guild.SetDB(db)
	for _, plugin := range ListPlugin() {
		plugin.Initialize(bot, db)
		log.WithFields(log.Fields{"plugin": plugin.GetName()}).Info("initialized plugin")
	}

	componentHandlers := make(map[string]func(*discordgo.Session, *discordgo.InteractionCreate))
	commandHandlers := make(map[string]func(*discordgo.Session, *discordgo.InteractionCreate))
	commands := make([]*discordgo.ApplicationCommand, 0, 2)

	// Add commands and handlers for the bot itself
	commands = append(commands, helpCommands...)
	for key, value := range helpCommandHandler {
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
	log.Debug("add bot handlers")
	bot.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			} else {
				log.WithField("command", i.ApplicationCommandData().Name).Warn("unhandled command")
			}
		case discordgo.InteractionMessageComponent:
			if h, ok := componentHandlers[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			} else {
				log.WithField("component", i.MessageComponentData().CustomID).Warn("unhandled component")
			}
		}
	})
	log.Debug("bot handlers added")

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
	log.Trace("--> discord.Bot.DeleteCommands")
	defer log.Trace("<-- discord.Bot.DeleteCommands")

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

	log.Debug("deleting old bot commands")
	_, err := bot.Session.ApplicationCommandBulkOverwrite(bot.appID, bot.guildID, nil)
	if err != nil {
		log.WithField("error", err).Fatal("failed to delete old commands")
	}
	log.Debug("old bot commands deleted")
}

// LoadCommands register all the commands. This implicitly will call the function added above that
// adds the command and component handlers for each plugin.
func (bot *Bot) LoadCommands(commands []*discordgo.ApplicationCommand) {
	log.Trace("--> discord.Bot.LoadCommands")
	defer log.Trace("<-- discord.Bot.LoadCommands")

	log.WithFields(log.Fields{"appID": bot.appID, "guildID": bot.guildID}).Debug("load new bot commands")
	_, err := bot.Session.ApplicationCommandBulkOverwrite(bot.appID, bot.guildID, commands)
	if err != nil {
		for _, command := range commands {
			log.WithFields(log.Fields{"name": command.Name, "description": command.Description}).Error("failed to load command")
		}
		log.WithFields(log.Fields{"appID": bot.appID, "guildID": bot.guildID, "commands": commands, "error": err}).Fatal("failed to load bot commands")

	}
	log.Info("new bot commands loaded")
}
