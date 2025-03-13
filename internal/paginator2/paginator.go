package paginator2

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Paginator represents a single paginator. It contains the the data to
// page through the user's data.
type Paginator struct {
	id           string
	title        string
	content      []*discordgo.MessageEmbedField
	cxpiry       time.Time
	itemsPerPage int
	currentPage  int
	config       *Config
	manager      *Manager
	channelID    string
	messageID    string
	ephemeral    bool
}

// NewPaginator creates a new paginator.
func NewPaginator(manager *Manager, title string, content []*discordgo.MessageEmbedField, itemsPerPage int, expiry time.Time) *Paginator {
	paginator := &Paginator{
		title:        title,
		content:      content,
		cxpiry:       expiry,
		currentPage:  0,
		itemsPerPage: itemsPerPage,
		config:       &defaultConfig,
		manager:      manager,
	}

	return paginator
}

// CreateMessage creates and sends a message withthe paginator's content.
func (p *Paginator) CreateMessage(s *discordgo.Session, channelID string, ephemeral ...bool) (*discordgo.Message, error) {
	if p.id == "" {
		p.id = fmt.Sprintf("%s-%d", channelID, time.Now().UnixNano())
		p.manager.Add(p)
	}
	p.channelID = channelID
	p.ephemeral = len(ephemeral) > 0 && ephemeral[0]
	var flags discordgo.MessageFlags
	if p.ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	embeds := []*discordgo.MessageEmbed{p.makeEmbed()}
	components := []discordgo.MessageComponent{p.createComponents()}
	message, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds:     embeds,
		Components: components,
		Flags:      flags,
	})
	if err != nil {
		return nil, err
	}
	p.messageID = message.ID
	return message, nil
}

// editMessage edits the current message sent by the paginator in a channel.
func (p *Paginator) editMessage(s *discordgo.Session) error {
	var flags discordgo.MessageFlags
	if p.ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	embeds := []*discordgo.MessageEmbed{p.makeEmbed()}
	components := []discordgo.MessageComponent{p.createComponents()}
	_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         p.messageID,
		Channel:    p.channelID,
		Embeds:     &embeds,
		Components: &components,
		Flags:      flags,
	})
	return err
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

// createComponents creates  the message components to be included in the
// message. It returns an action row that contains the buttons used to navigate
// through the paginator.
func (p *Paginator) createComponents() discordgo.MessageComponent {
	cfg := p.config.ButtonsConfig
	actionRow := discordgo.ActionsRow{}

	if cfg.First != nil {
		actionRow.Components = append(actionRow.Components, discordgo.Button{
			Label:    cfg.First.Label,
			Style:    cfg.First.Style,
			Disabled: p.currentPage == 0,
			Emoji:    cfg.First.Emoji,
			CustomID: p.formatCustomID("first"),
		})
	}
	if cfg.Back != nil {
		actionRow.Components = append(actionRow.Components, discordgo.Button{
			Label:    cfg.Back.Label,
			Style:    cfg.Back.Style,
			Disabled: p.currentPage == 0,
			Emoji:    cfg.Back.Emoji,
			CustomID: p.formatCustomID("back"),
		})
	}
	if cfg.Next != nil {
		actionRow.Components = append(actionRow.Components, discordgo.Button{
			Label:    cfg.Next.Label,
			Style:    cfg.Next.Style,
			Disabled: p.currentPage == p.pageCount()-1,
			Emoji:    cfg.Next.Emoji,
			CustomID: p.formatCustomID("next"),
		})
	}
	if cfg.Last != nil {
		actionRow.Components = append(actionRow.Components, discordgo.Button{
			Label:    cfg.Last.Label,
			Style:    cfg.Last.Style,
			Disabled: p.currentPage == p.pageCount()-1,
			Emoji:    cfg.Last.Emoji,
			CustomID: p.formatCustomID("last"),
		})
	}

	return actionRow
}

// formatCustomID formats the custom ID for the paginator buttons.
func (p *Paginator) formatCustomID(action string) string {
	return p.config.CustomIDPrefix + ":" + p.id + ":" + action
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
