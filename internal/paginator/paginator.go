package paginator

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/discord"
	log "github.com/sirupsen/logrus"
)

var (
	bot *discord.Bot
)

// SetBot sets the discord bot used to interact with the Discord servers.
func SetBot(b *discord.Bot) {
	bot = b
}

// Paginator represents a single paginator. It contains the the data to
// page through the user's data.
type Paginator struct {
	id           string
	title        string
	content      []*discordgo.MessageEmbedField
	idleWait     time.Duration
	expiry       time.Time
	itemsPerPage int
	currentPage  int
	config       *Config
	channelID    string
	interaction  *discordgo.Interaction
	ephemeral    bool
}

// NewPaginator creates a new paginator.
func NewPaginator(title string, itemsPerPage int, idleWait time.Duration, content []*discordgo.MessageEmbedField) *Paginator {
	paginator := &Paginator{
		title:        title,
		content:      content,
		idleWait:     idleWait,
		expiry:       time.Now().Add(idleWait),
		currentPage:  0,
		itemsPerPage: itemsPerPage,
		config:       &defaultConfig,
	}

	log.WithFields(log.Fields{"title": title, "itemsPerPage": itemsPerPage, "idleWait": idleWait, "content": len(content)}).Debug("creating new paginator")
	return paginator
}

// CreateMessage creates and sends a message withthe paginator's content.
func (p *Paginator) CreateMessage(s *discordgo.Session, i *discordgo.InteractionCreate, ephemeral ...bool) error {
	if p.id == "" {
		p.id = fmt.Sprintf("%s-%d", i.ChannelID, time.Now().UnixNano())
		manager.Add(p)
	}
	p.channelID = i.ChannelID
	p.interaction = i.Interaction
	p.ephemeral = len(ephemeral) > 0 && ephemeral[0]
	var flags discordgo.MessageFlags
	if p.ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	p.registerComponentHandlers()
	embeds := []*discordgo.MessageEmbed{p.makeEmbed()}
	components := []discordgo.MessageComponent{p.makeComponent()}
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     embeds,
			Components: components,
			Flags:      flags,
		},
	})
	if err != nil {
		log.WithFields(log.Fields{"paginator": p.id, "channel": i.ChannelID, "error": err}).Error("error sending paginated message")
		p.deregisterComponentHandlers()
		return err
	}
	log.WithFields(log.Fields{"paginator": p.id, "channel": i.ChannelID}).Debug("created paginated message")
	return nil
}

// editMessage edits the current message sent by the paginator in a channel.
func (p *Paginator) editMessage(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// var flags discordgo.MessageFlags
	// if p.ephemeral {
	// 	flags = discordgo.MessageFlagsEphemeral
	// }

	embeds := []*discordgo.MessageEmbed{p.makeEmbed()}
	components := []discordgo.MessageComponent{p.makeComponent()}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	_, err := s.InteractionResponseEdit(p.interaction, &discordgo.WebhookEdit{
		Embeds:     &embeds,
		Components: &components,
		// Flags:      flags,
	})
	if err != nil {
		log.WithFields(log.Fields{"paginator": p.id, "channel": p.channelID, "error": err}).Error("error editing paginated message")
		return err
	}

	log.WithFields(log.Fields{"paginator": p.id, "channel": p.channelID}).Debug("edited paginated message")
	return nil
}

// pageCount returns the number of pages in the paginator.
func (p *Paginator) pageCount() int {
	itemsPerPage := p.getItemsPerPage()
	pageCount := (len(p.content) + itemsPerPage - 1) / itemsPerPage
	return pageCount
}

// makeEmbed creates the message embed to be included for the current page.
func (p *Paginator) makeEmbed() *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Color:  p.config.EmbedColor,
		Title:  p.title,
		Fields: make([]*discordgo.MessageEmbedField, 0, p.getItemsPerPage()),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Page %d of %d", p.currentPage+1, p.pageCount()),
		},
	}
	start := p.currentPage * p.getItemsPerPage()
	end := min(start+p.getItemsPerPage(), len(p.content))
	embed.Fields = append(embed.Fields, p.content[start:end]...)
	return embed
}

// makeComponent creates  the message components to be included in the
// message. It returns an action row that contains the buttons used to navigate
// through the paginator.
func (p *Paginator) makeComponent() discordgo.MessageComponent {
	cfg := p.config.ButtonsConfig
	actionRow := discordgo.ActionsRow{}

	if cfg.First != nil {
		buttonID := p.id + ":first"
		actionRow.Components = append(actionRow.Components, discordgo.Button{
			Label:    cfg.First.Label,
			Style:    cfg.First.Style,
			Disabled: p.currentPage == 0,
			Emoji:    cfg.First.Emoji,
			CustomID: buttonID,
		})
	}
	if cfg.Back != nil {
		buttonID := p.id + ":back"
		actionRow.Components = append(actionRow.Components, discordgo.Button{
			Label:    cfg.Back.Label,
			Style:    cfg.Back.Style,
			Disabled: p.currentPage == 0,
			Emoji:    cfg.Back.Emoji,
			CustomID: buttonID,
		})
	}
	if cfg.Next != nil {
		buttonID := p.id + ":next"
		actionRow.Components = append(actionRow.Components, discordgo.Button{
			Label:    cfg.Next.Label,
			Style:    cfg.Next.Style,
			Disabled: p.currentPage == p.pageCount()-1,
			Emoji:    cfg.Next.Emoji,
			CustomID: buttonID,
		})
	}
	if cfg.Last != nil {
		buttonID := p.id + ":last"
		actionRow.Components = append(actionRow.Components, discordgo.Button{
			Label:    cfg.Last.Label,
			Style:    cfg.Last.Style,
			Disabled: p.currentPage == p.pageCount()-1,
			Emoji:    cfg.Last.Emoji,
			CustomID: buttonID,
		})
	}

	return actionRow
}

// registerComponentHandlers registers the component handlers for the paginator.
func (p *Paginator) registerComponentHandlers() {
	cfg := p.config.ButtonsConfig
	if cfg.First != nil {
		buttonID := p.id + ":first"
		bot.AddComponentHandler(buttonID, pageThroughItems)
	}
	if cfg.Back != nil {
		buttonID := p.id + ":back"
		bot.AddComponentHandler(buttonID, pageThroughItems)
	}
	if cfg.Next != nil {
		buttonID := p.id + ":next"
		bot.AddComponentHandler(buttonID, pageThroughItems)
	}
	if cfg.Last != nil {
		buttonID := p.id + ":last"
		bot.AddComponentHandler(buttonID, pageThroughItems)
	}
	log.WithFields(log.Fields{"paginator": p.id}).Debug("registered component handlers")
}

// deregisterComponentHandlers deregisters the component handlers for the paginator.
func (p *Paginator) deregisterComponentHandlers() {
	cfg := p.config.ButtonsConfig
	if cfg.First != nil {
		buttonID := p.id + ":first"
		bot.RemoveComponentHandler(buttonID)
	}
	if cfg.Back != nil {
		buttonID := p.id + ":back"
		bot.RemoveComponentHandler(buttonID)
	}
	if cfg.Next != nil {
		buttonID := p.id + ":next"
		bot.RemoveComponentHandler(buttonID)
	}
	if cfg.Last != nil {
		buttonID := p.id + ":last"
		bot.RemoveComponentHandler(buttonID)
	}
	log.WithFields(log.Fields{"paginator": p.id}).Debug("deregistered component handlers")
}

// itemsPerPage returns the number of items per page. If the
// ItemsPerPage field is 0, it returns the default number of items
// per page.
func (p *Paginator) getItemsPerPage() int {
	if p.itemsPerPage == 0 {
		return p.config.DefaultItemsPerPage
	}
	return p.itemsPerPage
}

// hasExpired returns true if the paginator has expired.
func (p *Paginator) hasExpired() bool {
	return !p.expiry.IsZero() && p.expiry.Before(time.Now())
}

// pageThroughItems is called when a page button is selected in a paginated message.
func pageThroughItems(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ids := strings.Split(i.Interaction.MessageComponentData().CustomID, ":")
	paginatorID, action := ids[0], ids[1]

	manager.mutex.Lock()
	paginator, ok := manager.paginators[paginatorID]
	manager.mutex.Unlock()
	if !ok {
		return
	}

	switch action {
	case "first":
		paginator.currentPage = 0

	case "back":
		paginator.currentPage--

	case "next":
		paginator.currentPage++

	case "last":
		paginator.currentPage = paginator.pageCount() - 1
	}

	paginator.expiry = time.Now().Add(paginator.idleWait)
	paginator.editMessage(s, i)
}
