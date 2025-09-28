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

	probAnyOrderRWB := getProabilityOfAnyOrderRedWhiteBlue(sm)
	probabilities = append(probabilities, *probAnyOrderRWB)

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

	var payoutAmount float64
	for _, payout := range sm.PayoutTable {
		if strings.Contains(payout.Win[0], slots.TwoConsecutiveSymbols) {
			payoutAmount = payout.Payout
			break
		}
	}

	probability := (float64(numMatches) / float64((nymPossibilities)))

	return &PayoutProbability{
		Spin:        []string{"two consecutive non-Spell symbols"},
		Payout:      slots.PayoutAmount{Bet: 1, Payout: payoutAmount},
		Probability: probability,
		NumMatches:  numMatches,
		Return:      (float64(payoutAmount) / float64(1)) * (probability) * 100.0,
	}
}

func getProabilityOfAnyOrderRedWhiteBlue(sm *slots.SlotMachine) *PayoutProbability {
	nymPossibilities := 1
	for _, reel := range sm.LookupTable {
		nymPossibilities *= len(reel)
	}

	numMatches := 0
	for _, symbol1 := range sm.LookupTable[0] {
		for _, symbol2 := range sm.LookupTable[1] {
			for _, symbol3 := range sm.LookupTable[2] {
				switch {
				case symbol1.Name == "Archer" || symbol1.Name == "AQ":
					switch {
					case symbol2.Name == "Wizard" || symbol2.Name == "GW":
						if symbol3.Name == "Barbarian" || symbol3.Name == "BK" {
							numMatches++
						}
					case symbol2.Name == "Barbarian" || symbol2.Name == "BK":
						if symbol3.Name == "Wizard" || symbol3.Name == "GW" {
							numMatches++
						}
					}
				case symbol1.Name == "Wizard" || symbol1.Name == "GW":
					switch {
					case symbol2.Name == "Archer" || symbol2.Name == "AQ":
						if symbol3.Name == "Barbarian" || symbol3.Name == "BK" {
							numMatches++
						}
					case symbol2.Name == "Barbarian" || symbol2.Name == "BK":
						if symbol3.Name == "Archer" || symbol3.Name == "AQ" {
							numMatches++
						}
					}
				case symbol1.Name == "Barbarian" || symbol1.Name == "BK":
					switch {
					case symbol2.Name == "Archer" || symbol2.Name == "AQ":
						if symbol3.Name == "Wizard" || symbol3.Name == "GW" {
							numMatches++
						}
					case symbol2.Name == "Wizard" || symbol2.Name == "GW":
						if symbol3.Name == "Archer" || symbol3.Name == "AQ" {
							numMatches++
						}
					}
				}
			}
		}
	}
	var payoutAmount float64
	for _, payout := range sm.PayoutTable {
		if strings.Contains(payout.Win[0], slots.AnyOrderRWB) {
			payoutAmount = payout.Payout
			break
		}
	}

	probability := (float64(numMatches) / float64((nymPossibilities)))

	return &PayoutProbability{
		Spin:        []string{"any order AQ/Archer, GW/Wizard, BK/Barbarian"},
		Payout:      slots.PayoutAmount{Bet: 1, Payout: payoutAmount},
		Probability: probability,
		NumMatches:  numMatches,
		Return:      (float64(payoutAmount) / float64(1)) * (probability) * 100.0,
	}
}
