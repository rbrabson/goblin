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

// PageCount returns the number of pages in the paginator.
func (p *Paginator) PageCount() int {
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
			Text: fmt.Sprintf("Page %d/%d", p.currentPage+1, p.PageCount()),
		},
	}
	start := p.currentPage * p.itemsPerPage()
	end := min(start+p.itemsPerPage(), len(p.Content))
	embed.Fields = append(embed.Fields, p.Content[start:end]...)
	return embed
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
