package discord

import (
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

const (
	botIntents = discordgo.IntentGuilds |
		discordgo.IntentGuildMessages |
		discordgo.IntentDirectMessages |
		discordgo.IntentGuildEmojis
	botVersion = "0.1.0"
)

// Bot is a Discord bot which is capable of running multiple services, each of which
// implement various commands.
type Bot struct {
	Session *discordgo.Session
	appID   string
	guildID string
	timer   chan int
}

// NewBot creates a nbew Discord bot that can run Discord commands.
func NewBot() *Bot {
	log.Trace("--> discord.NewBot")

	// Use environment variables to pass in sensitive data used to identify the bot
	godotenv.Load()
	appID := os.Getenv("DISCORD_APP_ID")
	token := os.Getenv("DISCORD_BOT_TOKEN")
	guildID := os.Getenv("DISCORD_GUILD_ID")
	deleteSlashCommands := GetenvBool("DISCORD_DELETE_SLASH_COMMANDS")

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
		log.Info("bot is up!")
	})

	for _, plugin := range ListPlugin() {
		plugin.Initialize(bot)
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
			}
		case discordgo.InteractionMessageComponent:
			if h, ok := componentHandlers[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			}
		}
	})

	if deleteSlashCommands {
		bot.DeleteCommands()
	}
	bot.LoadCommands(commands)

	log.Trace("<-- discord.NewBot")
	return bot
}

// DeleteCommands deletes the current set of slash commands. This can be useful when developing
// a new bot and the set of loaded slash commands changes.
func (bot *Bot) DeleteCommands() {
	log.Trace("--> discord.Bot.DeleteCommands")

	_, err := bot.Session.ApplicationCommandBulkOverwrite(bot.appID, bot.guildID, nil)
	if err != nil {
		log.WithField("error", err).Fatal("failed to delete old commands")
	}
	log.Info("old bot commands deleted")

	log.Trace("<-- discord.Bot.DeleteCommands")
}

// LoadCommands register all the commands. This implicitly will call the function added above that
// adds the command and component handlers for each plugin.
func (bot *Bot) LoadCommands(commands []*discordgo.ApplicationCommand) {
	log.Trace("--> discord.Bot.LoadCommands")

	_, err := bot.Session.ApplicationCommandBulkOverwrite(bot.appID, bot.guildID, commands)
	if err != nil {
		log.WithField("error", err).Fatal("failed to load bot commands")
	}
	log.Info("new bot commands loaded")

	log.Trace("<-- discord.Bot.LoadCommands")
}
