package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/game/slots"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type PayoutProbability struct {
	Spin        []string
	Payout      slots.PayoutAmount
	Probability float64
	NumMatches  int
	Return      float64
}

func main() {
	godotenv.Load(".env")

	p := message.NewPrinter(language.AmericanEnglish)

	sm := slots.NewSlotMachine("xxxx")
	payoutTable := sm.PayoutTable
	lookupTable := sm.LookupTable

	nymPossibilities := 1
	for _, reel := range lookupTable.Reels {
		nymPossibilities *= len(reel)
	}

	probabilities := make([]PayoutProbability, 0, len(payoutTable.Payouts))

	for _, payout := range payoutTable.Payouts {
		numMatches := 0
		for i, winningSymbols := range payout.Win {
			matchingSymbolsOnReel := getMatchingSymbols(winningSymbols, &lookupTable.Reels[i])
			if numMatches == 0 {
				numMatches = matchingSymbolsOnReel
			} else {
				numMatches *= matchingSymbolsOnReel
			}
		}
		probability := (float64(numMatches) / float64((nymPossibilities))) * 100.0
		probabilities = append(probabilities, PayoutProbability{
			Spin:        payout.Win,
			Payout:      payout,
			Probability: probability,
			NumMatches:  numMatches,
			Return:      (float64(payout.Payout) / float64(payout.Bet)) * (probability / 100.0) * 100.0,
		})
	}

	totalWinProb := 0.0
	for _, prob := range probabilities {
		totalWinProb += prob.Probability
	}
	totalReturnPercentage := 0.0
	for _, prob := range probabilities {
		totalReturnPercentage += prob.Return
	}

	p.Printf("Total Win Probability: %v%%\n", totalWinProb*100.0)

	p.Println("Payout Probabilities:")
	for _, prob := range probabilities {
		p.Printf("Spin: %v, NumMatches: %d, Probability: %.4f%%, Return: %.4f%%\n", prob.Spin, prob.NumMatches, prob.Probability, prob.Return)
	}
	p.Printf("Total Win Probability: %.2f%%\n", totalWinProb)
	p.Printf("Total Return Percentage: %.2f%%\n", totalReturnPercentage)

	// printReels(lookupTable.Reels)
}

func getMatchingSymbols(winningSymbols string, reel *slots.Reel) int {
	symbols := strings.Split(winningSymbols, " or ")
	matchingSymbols := 0
	for _, symbol := range *reel {
		if slices.Contains(symbols, symbol.Name) {
			matchingSymbols++
		}
	}

	return matchingSymbols
}

func printReels(reels []slots.Reel) {
	for i := range len(reels[0]) {
		fmt.Printf("[%d]\t", i+1)
		for _, reel := range reels {
			print(reel[i].Name, "\t")
		}
		print("\n")
	}
	println()
}
