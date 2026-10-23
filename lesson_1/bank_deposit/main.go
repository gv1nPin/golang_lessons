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
	initialDeposit := int64(100000 * 100)
	balance := initialDeposit
	annualRate := 12.0
	months := 12

	monthlyRate := (annualRate / 100) / 12
	monthCount := 0

	fmt.Printf("Начальный депозит: %.2f под %.1f%% годовых с ежемесячной капитализацией:\n\n", balance, annualRate)

	for month := 1; month <= months; month++ {
		interestInKopecks := int64((float64(balance) * monthlyRate) + 0.5)
		balance += interestInKopecks
		fmt.Printf("месяц %2d: %.2f\n", months, float64(balance)/100)

		for balance < (initialDeposit * 2) {
			interestInKopecks := int64((float64(balance) * monthlyRate) + 0.5)
			balance += interestInKopecks
			monthCount++
		}
	}
	rubles := balance / 100
	kopecks := balance % 100
	fmt.Printf("Вклад удвоится через %d мес. Баланс: %d.%02d\n", monthCount, rubles, kopecks)
}
