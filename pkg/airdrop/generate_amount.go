package airdrop

import (
	"fmt"
	math "math"
	"math/rand"
	"strconv"
	"strings"
)

type Threshold struct {
	Balance    float64
	AirdropAmt float64
}

func GenerateAirdropAmount(bnbBalance float64, thresholds []Threshold) float64 {
	var lastThreshold Threshold
	for _, th := range thresholds {
		if bnbBalance <= th.Balance {
			if lastThreshold.Balance == 0 {
				return th.AirdropAmt
			}
			proportion := (bnbBalance - lastThreshold.Balance) / (th.Balance - lastThreshold.Balance)
			return lastThreshold.AirdropAmt + proportion*(th.AirdropAmt-lastThreshold.AirdropAmt)
		}
		lastThreshold = th
	}
	return lastThreshold.AirdropAmt
}

func roundToPrecision(value float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	return math.Round(value*pow) / pow
}

func RandomizeDecimalCount(originalFloat float64) float64 {
	floatStr := fmt.Sprintf("%f", originalFloat)
	decimalPointIndex := len(floatStr) - strings.IndexFunc(strconv.FormatFloat(originalFloat, 'f', -1, 64), func(r rune) bool {
		return r == '.'
	}) - 1
	newDecimalCount := rand.Intn(decimalPointIndex + 1)
	return roundToPrecision(originalFloat, newDecimalCount)
}
