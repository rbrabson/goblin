package main

import (
	"fmt"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/game/slots"
	rslots "github.com/rbrabson/slots"
)

type PayoutProbability struct {
	Spin        []string
	Payout      rslots.PayoutAmount
	Probability float64
	NumMatches  int
	Return      float64
	Message     string
}

func main() {
	godotenv.Load(".env")

	sm := rslots.NewSlotMachine(
		rslots.WithLookupTable(slots.GetLookupTable()),
		rslots.WithPayoutTable(slots.GetPayoutTable()),
	)

	nymPossibilities := 1
	for _, reel := range sm.LookupTable {
		nymPossibilities *= len(reel)
	}

	probabilities := make([]PayoutProbability, 0, len(sm.PayoutTable))
	for _, payout := range sm.PayoutTable {
		payoutProbability := getProbabilityOfWin(&payout, sm)
		probabilities = append(probabilities, *payoutProbability)
	}

	totalWinProb := 0.0
	totalReturn := 0.0
	for _, prob := range probabilities {
		totalWinProb += prob.Probability
		totalReturn += prob.Return
	}

	fmt.Println("Spin, Matches, Payout, Probability, Return")
	for _, prob := range probabilities {
		if prob.NumMatches != 0 {
			payoutStr := strconv.FormatFloat(prob.Payout.Payout, 'f', -1, 64)
			fmt.Printf("%s, %d, %d:%s, %.4f%%, %.4f%%\n", prob.Message, prob.NumMatches, prob.Payout.Bet, payoutStr, prob.Probability, prob.Return)
		}
	}

	fmt.Printf("\nWin,,, %.2f%%, %.2f%%\n", totalWinProb, totalReturn)
}

func getProbabilityOfWin(payout *rslots.PayoutAmount, sm *rslots.SlotMachine) *PayoutProbability {
	nymPossibilities := 1
	for _, reel := range sm.LookupTable {
		nymPossibilities *= len(reel)
	}

	numMatches := 0

	for _, symbol1 := range sm.LookupTable[0] {
		for _, symbol2 := range sm.LookupTable[1] {
			for _, symbol3 := range sm.LookupTable[2] {
				payoutAmount := payout.GetPayoutAmount(1, []string{symbol1, symbol2, symbol3})
				if payoutAmount > 0 {
					numMatches++
				}
			}
		}
	}

	bet := float64(payout.Bet)
	payoutAmount := float64(payout.Payout)
	probability := (float64(numMatches) / float64((nymPossibilities)))

	return &PayoutProbability{
		Spin:        payout.Win,
		Payout:      rslots.PayoutAmount{Bet: payout.Bet, Payout: payout.Payout},
		Probability: probability * 100.0,
		NumMatches:  numMatches,
		Message:     payout.Message,
		Return:      (payoutAmount - bet) / bet * probability * 100.0,
	}
}
