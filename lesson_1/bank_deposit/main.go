package main

import "fmt"

const (
	initialDeposit = 100000.0
	annualRate     = 12.0
	monthsInYear   = 12
)

// float64 хранит числа в двоичном виде и не может точно представить
// большинство десятичных дробей (0.1, 0.01 и т.д.). Из-за этого появляются
// мелкие ошибки округления, которые накапливаются в суммах и ломают сравнения
// (0.1+0.2 != 0.3). Для денег используют целые числа в минимальных единицах
// (копейки в int64) или десятичные типы (decimal) с явным правилом округления.
// float64 оставлен намеренно, чтобы совпасть с примерами из задания.
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
