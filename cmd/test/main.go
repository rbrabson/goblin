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

	// lookupTable := slots.GetLookupTable("1234567890")
	// bytes, _ := json.MarshalIndent(lookupTable.Reels, "", "  ")
	// reels := string(bytes)
	// fmt.Println(reels)

	spin := []string{"Archer Queen", "Archer Queen", "Archer Queen"}
	payoutTable := slots.GetPayoutTable("1234567890")
	winAmount := payoutTable.GetPayoutAmount(300, spin)
	p.Println(winAmount)

	// symbols := slots.GetSymbols("1234567890")
	// bytes, _ := json.MarshalIndent(symbols, "", "  ")
	// symbolsStr := string(bytes)
	// fmt.Println(symbolsStr)
}
