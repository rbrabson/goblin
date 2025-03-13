package paginator2

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Paginator represents a single paginator. It contains the the data to
// page through the user's data.
type Paginator struct {
	ID           string
	Title        string
	Content      []*discordgo.MessageEmbedField
	Expiry       time.Time
	ItemsPerPage int
	currentPage  int
	config       *Config
}

// NewPaginator creates a new paginator.
func NewPaginator(id string, title string, content []*discordgo.MessageEmbedField, expiry time.Time) *Paginator {
	paginator := &Paginator{
		ID:          id,
		Title:       title,
		Content:     content,
		Expiry:      expiry,
		currentPage: 0,
		config:      &defaultConfig,
	}
	return paginator
}

// pageCount returns the number of pages in the paginator.
func (p *Paginator) pageCount() int {
	itemsPerPage := p.itemsPerPage()
	pageCount := (len(p.Content) + itemsPerPage - 1) / itemsPerPage
	return pageCount
}

// makeEmbed creates the message embed to be included for the current page.
func (p *Paginator) makeEmbed() *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Color:  p.config.EmbedColor,
		Title:  p.Title,
		Fields: make([]*discordgo.MessageEmbedField, 0, p.itemsPerPage()),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Page %d/%d", p.currentPage+1, p.pageCount()),
		},
	}
	start := p.currentPage * p.itemsPerPage()
	end := min(start+p.itemsPerPage(), len(p.Content))
	embed.Fields = append(embed.Fields, p.Content[start:end]...)
	return embed
}

// createComponents creates  the message components to be included in the
// message. It contains the buttons used to navigate through the paginator.
func (p *Paginator) createComponents(paginator *Paginator) discordgo.MessageComponent {
	cfg := p.config.ButtonsConfig
	actionRow := discordgo.ActionsRow{}

	if cfg.First != nil {
		actionRow.Components = append(actionRow.Components, discordgo.Button{
			Label:    cfg.First.Label,
			Style:    cfg.First.Style,
			Disabled: paginator.currentPage == 0,
			Emoji:    cfg.First.Emoji,
			CustomID: p.formatCustomID("first"),
		})
	}
	if cfg.Back != nil {
		actionRow.Components = append(actionRow.Components, discordgo.Button{
			Label:    cfg.Back.Label,
			Style:    cfg.Back.Style,
			Disabled: paginator.currentPage == 0,
			Emoji:    cfg.Back.Emoji,
			CustomID: p.formatCustomID("back"),
		})
	}

	if cfg.Next != nil {
		actionRow.Components = append(actionRow.Components, discordgo.Button{
			Label:    cfg.Next.Label,
			Style:    cfg.Next.Style,
			Disabled: paginator.currentPage == paginator.pageCount()-1,
			Emoji:    cfg.Next.Emoji,
			CustomID: p.formatCustomID("next"),
		})
	}
	if cfg.Last != nil {
		actionRow.Components = append(actionRow.Components, discordgo.Button{
			Label:    cfg.Last.Label,
			Style:    cfg.Last.Style,
			Disabled: paginator.currentPage == paginator.pageCount()-1,
			Emoji:    cfg.Last.Emoji,
			CustomID: p.formatCustomID("last"),
		})
	}

	return actionRow
}

// formatCustomID formats the custom ID for the paginator buttons.
func (p *Paginator) formatCustomID(action string) string {
	return p.config.CustomIDPrefix + ":" + p.ID + ":" + action
}

// itemsPerPage returns the number of items per page. If the
// ItemsPerPage field is 0, it returns the default number of items
// per page.
func (p *Paginator) itemsPerPage() int {
	if p.ItemsPerPage == 0 {
		return p.config.DefaultItemsPerPage
	}
	return p.ItemsPerPage
}
