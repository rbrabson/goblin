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

	lookupTable := slots.GetLookupTable("1234567890")
	payoutTable := slots.GetPayoutTable("1234567890")

	spin := lookupTable.Spin()

	betAmount := 300
	winAmount := payoutTable.GetPayoutAmount(betAmount, spin.Spins[spin.WinIndex])

	p.Printf("Win: %v\n", spin.Spins[spin.WinIndex])
	p.Printf("Bet: %d\n", betAmount)
	p.Printf("Win: %d\n", winAmount)

	for i, spin := range spin.Spins {
		p.Printf("Spin[%d]: %v\n", i, spin)
	}
}
