package main

import (
	"fmt"
	"math"
)

func calculateSimpleInterest(amount float64, rate float64, days int) float64 {
	const daysInYear = 365.0
	return (amount * rate * float64(days)) / (daysInYear * 100)
}

func calculateCompoundInterest(amount float64, rate float64, months int) float64 {
	periodRate := rate / (12 * 100)
	totalAmount := amount * math.Pow(1+periodRate, float64(months))
	return totalAmount - amount
}

func main() {
	deposit := 100000.0
	annualRate := 12.0
	months := 12

	montlyRate := (annualRate / 100) / 12

	fmt.Printf("Начальный депозит: %.2f под %.1f%% годовых с ежемесячной капитализацией:\n\n", deposit, annualRate)

	for month := 1; month <= months; month++ {
		deposit += deposit * montlyRate
		fmt.Printf("месяц %2d: %.2f\n", months, deposit)
	}
}
