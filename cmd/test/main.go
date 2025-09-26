package main

import (
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
	Return      float64
}

func main() {
	godotenv.Load(".env")

	p := message.NewPrinter(language.AmericanEnglish)

	sm := slots.NewSlotMachine("xxxx")
	payoutTable := sm.PayoutTable
	lookupTable := sm.LookupTable

	probabilities := make([]PayoutProbability, 0, len(payoutTable.Payouts))

	for _, payout := range payoutTable.Payouts {
		probability := 0.0
		for i, winningSymbols := range payout.Win {
			reelProbability := getProbability(winningSymbols, &lookupTable.Reels[i])
			if probability == 0.0 {
				probability = reelProbability
			} else {
				probability *= reelProbability
			}
		}
		probabilities = append(probabilities, PayoutProbability{
			Spin:        payout.Win,
			Payout:      payout,
			Probability: probability * 100.0,
			Return:      (float64(payout.Payout) / float64(payout.Bet)) * probability * 100.0,
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
		p.Printf("Spin: %v, Payout: %v, Probability: %.4f%%, Return: %.6f\n", prob.Spin, prob.Payout.Payout, prob.Probability, prob.Return/100.0)
	}
	p.Printf("Total Win Probability: %f%%\n", totalWinProb)
	p.Printf("Total Return Percentage: %.4f%%\n", totalReturnPercentage)
}

func getProbability(winningSymbols string, reel *slots.Reel) float64 {
	symbols := strings.Split(winningSymbols, " or ")
	matchingSymbols := 0
	for _, symbol := range *reel {
		if slices.Contains(symbols, symbol.Name) {
			matchingSymbols++
		}
	}

	return float64(matchingSymbols) / float64(len(*reel))
}
