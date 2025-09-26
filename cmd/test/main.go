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
	NumMatches  int
	Return      float64
}

func main() {
	godotenv.Load(".env")

	p := message.NewPrinter(language.AmericanEnglish)

	sm := slots.GetSlotMachine()
	payoutTable := sm.PayoutTable
	lookupTable := sm.LookupTable

	nymPossibilities := 1
	for _, reel := range lookupTable {
		nymPossibilities *= len(reel)
	}

	probabilities := make([]PayoutProbability, 0, len(payoutTable))

	for _, payout := range payoutTable {
		numMatches := 0
		for i, winningSymbols := range payout.Win {
			matchingSymbolsOnReel := getMatchingSymbols(winningSymbols, &lookupTable[i])
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

	probTwoNonBlank := getProbabilityOfTwoConsecutiveSymbols(sm)
	probabilities = append(probabilities, *probTwoNonBlank)

	totalWinProb := 0.0
	for _, prob := range probabilities {
		totalWinProb += prob.Probability
	}
	totalReturnPercentage := 0.0
	for _, prob := range probabilities {
		totalReturnPercentage += prob.Return
	}

	for _, prob := range probabilities {
		if prob.NumMatches != 0 {
			spin := "[" + strings.Join(prob.Spin, ", ") + "]"
			p.Printf("%s: NumMatches: %d, Probability: %.4f%%, Return: %.4f%%\n", spin, prob.NumMatches, prob.Probability, prob.Return)
		}
	}

	p.Printf("Total Win Probability: %.2f%%\n", totalWinProb)
	p.Printf("Total Return Percentage: %.2f%%\n", totalReturnPercentage)
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

func getProbabilityOfTwoConsecutiveSymbols(sm *slots.SlotMachine) *PayoutProbability {
	nymPossibilities := 1
	for _, reel := range sm.LookupTable {
		nymPossibilities *= len(reel)
	}

	numMatches := 0
	for _, symbol1 := range sm.LookupTable[0] {
		for _, symbol2 := range sm.LookupTable[1] {
			for _, symbol3 := range sm.LookupTable[2] {
				if (symbol1.Name != "Spell" && symbol1.Name == symbol2.Name && symbol1.Name != symbol3.Name) ||
					(symbol1.Name != symbol2.Name && symbol2.Name != "Spell" && symbol2.Name == symbol3.Name) {
					numMatches++
				}
			}
		}
	}

	probability := (float64(numMatches) / float64((nymPossibilities))) * 100.0

	return &PayoutProbability{
		Spin:        []string{slots.TwoConsecutiveSymbols},
		Payout:      slots.PayoutAmount{Bet: 1, Payout: 1},
		Probability: probability,
		NumMatches:  numMatches,
		Return:      (float64(1) / float64(1)) * (probability / 100.0) * 100.0,
	}
}
