package main

import (
	"encoding/json"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/game/slots"
)

func main() {
	godotenv.Load(".env")

	// lookupTable := slots.GetLookupTable("1234567890")
	// bytes, _ := json.MarshalIndent(lookupTable.Reels, "", "  ")
	// reels := string(bytes)
	// fmt.Println(reels)

	payoutTable := slots.GetPayoutTable("1234567890")
	bytes, _ := json.MarshalIndent(payoutTable.Payouts, "", "  ")
	paylines := string(bytes)
	fmt.Println(paylines)

	// symbols := slots.GetSymbols("1234567890")
	// bytes, _ := json.MarshalIndent(symbols, "", "  ")
	// symbolsStr := string(bytes)
	// fmt.Println(symbolsStr)
}
