package discmsg

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// GetPrinter returns a printer for the given locale of the user initiating the message.
func GetPrinter(tag language.Tag) *message.Printer {
	return message.NewPrinter(tag)
}
