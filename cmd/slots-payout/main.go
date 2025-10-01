package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/game/slots"
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

	sm := slots.GetSlotMachine()
	payoutTable := sm.PayoutTable
	lookupTable := sm.LookupTable

	nymPossibilities := 1
	for _, reel := range lookupTable {
		nymPossibilities *= len(reel)
	}

	probabilities := make([]PayoutProbability, 0, len(payoutTable))

	for _, payout := range payoutTable {
		payoutProbability := getProbabilityOfWin(&payout, sm)
		probabilities = append(probabilities, *payoutProbability)
	}

	totalWinProb := 0.0
	for _, prob := range probabilities {
		totalWinProb += prob.Probability
	}

	fmt.Println("Spin, Matches, Payout, Probability, Return")
	for _, prob := range probabilities {
		if prob.NumMatches != 0 {
			payoutStr := strconv.FormatFloat(prob.Payout.Payout, 'f', -1, 64)
			spin := "[" + strings.Join(prob.Spin, " | ") + "]"
			fmt.Printf("%s, %d, %d:%s, %.4f%%, %.4f%%\n", spin, prob.NumMatches, prob.Payout.Bet, payoutStr, prob.Probability, prob.Return)
		}
	}

	var totalBets, totalWins, totalReturn int
	for _, symbol1 := range sm.LookupTable[0] {
		for _, symbol2 := range sm.LookupTable[1] {
			for _, symbol3 := range sm.LookupTable[2] {
				totalBets += 1
				payout, _ := sm.PayoutTable.GetPayoutAmount(1, []slots.Symbol{symbol1, symbol2, symbol3})
				totalReturn += payout
				if payout > 0 {
					totalWins += 1
				}
			}
		}
	}
	totalReturnPercentage := (float64(totalReturn) / float64(totalBets)) * 100.0
	fmt.Printf("\nWin,,, %.2f%%, %.2f%%\n", totalWinProb, totalReturnPercentage)
}

func getProbabilityOfWin(payout *slots.PayoutAmount, sm *slots.SlotMachine) *PayoutProbability {
	nymPossibilities := 1
	for _, reel := range sm.LookupTable {
		nymPossibilities *= len(reel)
	}

	numMatches := 0
	for _, symbol1 := range sm.LookupTable[0] {
		for _, symbol2 := range sm.LookupTable[1] {
			for _, symbol3 := range sm.LookupTable[2] {
				_, msg := sm.PayoutTable.GetPayoutAmount(1, []slots.Symbol{symbol1, symbol2, symbol3})
				if msg == payout.Message {
					numMatches++
				}
			}
		}
	}

	bet := payout.Bet
	payoutAmount := payout.Payout
	probability := (float64(numMatches) / float64((nymPossibilities)))

	return &PayoutProbability{
		Spin:        payout.Win,
		Payout:      slots.PayoutAmount{Bet: bet, Payout: payoutAmount},
		Probability: probability * 100.0,
		NumMatches:  numMatches,
		Return:      (float64(payoutAmount) / float64(bet)) * probability * 100.0,
	}
}
