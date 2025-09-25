package main

import (
	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/game/slots"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func main() {
	godotenv.Load(".env")

	p := message.NewPrinter(language.AmericanEnglish)

	betAmount := 300
	sm := slots.NewSlotMachine(slots.DummyGuildID)
	spin := sm.Spin(betAmount)

	p.Printf("Bet: %d\n", betAmount)
	p.Printf("Win: %d\n", spin.Payout)
	p.Printf("Next: %v\n", spin.NextLine)
	p.Printf("Payline: %v\n", spin.Payline)
	p.Printf("Previous: %v\n", spin.PreviousLine)
}
