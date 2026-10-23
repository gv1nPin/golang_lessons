package main

import (
	"fmt"
	"math"
)

func calculateSimpleInterest(amount float64, rate float64, days int) float64 {
	const daysInYear = 365.0
	return (amount * rate * float64(days)) / (daysInYear * 100)
}

func main() {

}
