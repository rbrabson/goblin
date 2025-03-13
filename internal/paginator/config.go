package paginator

import (
	"github.com/bwmarrin/discordgo"
)

const (
	defaultItemsPerPage = 5
)

// defaultConfig is the default configuration used by the paginator.
var defaultConfig = Config{
	ButtonsConfig: ButtonsConfig{
		First: &ComponentOptions{
			Emoji: &discordgo.ComponentEmoji{
				Name: "⏮",
			},
			Style: discordgo.PrimaryButton,
		},
		Back: &ComponentOptions{
			Emoji: &discordgo.ComponentEmoji{
				Name: "◀",
			},
			Style: discordgo.PrimaryButton,
		},
		Next: &ComponentOptions{
			Emoji: &discordgo.ComponentEmoji{
				Name: "▶",
			},
			Style: discordgo.PrimaryButton,
		},
		Last: &ComponentOptions{
			Emoji: &discordgo.ComponentEmoji{
				Name: "⏭",
			},
			Style: discordgo.PrimaryButton,
		},
	},
	CustomIDPrefix:      "paginator",
	EmbedColor:          0x4c50c1,
	DefaultItemsPerPage: defaultItemsPerPage,
}

// Config is the configuration used by the paginator.
type Config struct {
	ButtonsConfig           ButtonsConfig
	NotYourPaginatorMessage string
	CustomIDPrefix          string
	EmbedColor              int
	DefaultItemsPerPage     int
}

// ButtonsConfig are the buttons used to navigate through the paginator.
type ButtonsConfig struct {
	First *ComponentOptions
	Back  *ComponentOptions
	Next  *ComponentOptions
	Last  *ComponentOptions
}

// ComponentOptions are the options used to create a pagination button.
type ComponentOptions struct {
	Emoji *discordgo.ComponentEmoji
	Label string
	Style discordgo.ButtonStyle
}

type ConfigOpt func(config *Config)

// Apply applies the given RequestOpt(s) to the RequestConfig & sets the context if none is set
func (c *Config) Apply(opts []ConfigOpt) {
	for _, opt := range opts {
		opt(c)
	}
}
