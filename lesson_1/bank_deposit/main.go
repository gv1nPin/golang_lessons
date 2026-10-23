package main

import "fmt"

const (
	initialDeposit = 100000.0
	annualRate     = 12.0
	monthsInYear   = 12
)

func main() {
	monthlyRate := annualRate / monthsInYear / 100 // 0.01
	balance := initialDeposit
	for month := 1; month <= monthsInYear; month++ {
		balance += balance * monthlyRate
		fmt.Printf("месяц %2d: %.2f\n", month, balance)
	}

	balance = initialDeposit
	months := 0
	for balance < initialDeposit*2 {
		balance += balance * monthlyRate
		months++
	}
	fmt.Printf("Вклад удвоится через %d мес. Баланс: %.2f\n", months, balance)
}
