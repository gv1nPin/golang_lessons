package main

import (
	"errors"
	"fmt"
)

var (
	ErrInsufficientFunds = errors.New("недостаточно средств")
	ErrInvalidAmount     = errors.New("сумма должна быть положительной")
)

// formatRub переводит копейки в строку вида "1199.50".
func formatRub(kopecks int64) string {
	sign := ""
	if kopecks < 0 {
		sign = "-"
		kopecks = -kopecks
	}
	return fmt.Sprintf("%s%d.%02d", sign, kopecks/100, kopecks%100)
}

// NewAccount создаёт счёт и возвращает три замыкания над общим балансом.
// Деньги хранятся в копейках (int64), чтобы не терять точность, как у float64.
// Каждый вызов создаёт свою переменную current, поэтому счета независимы.
//
// deposit и withdraw ведут себя одинаково: при некорректной сумме (0 или
// отрицательной) возвращают ErrInvalidAmount и баланс не меняют.
func NewAccount(initial int64) (deposit func(int64) error, withdraw func(int64) error, balance func() int64) {
	current := initial

	deposit = func(amount int64) error {
		if amount <= 0 {
			return ErrInvalidAmount
		}
		current += amount
		return nil
	}

	withdraw = func(amount int64) error {
		if amount <= 0 {
			return ErrInvalidAmount
		}
		if amount > current {
			return fmt.Errorf("%w (баланс %s)", ErrInsufficientFunds, formatRub(current))
		}
		current -= amount
		return nil
	}

	balance = func() int64 {
		return current
	}

	return deposit, withdraw, balance
}

func main() {
	depositA, withdrawA, balanceA := NewAccount(100000) // 1000.00 ₽
	_, _, balanceB := NewAccount(0)

	fmt.Printf("Счёт A: %s\n", formatRub(balanceA()))

	if err := depositA(50000); err != nil { // 500.00 ₽
		fmt.Printf("Пополнение %s: %v\n", formatRub(50000), err)
	} else {
		fmt.Printf("Пополнение %s -> %s\n", formatRub(50000), formatRub(balanceA()))
	}

	if err := withdrawA(200000); err != nil { // 2000.00 ₽
		fmt.Printf("Снятие %s: %v\n", formatRub(200000), err)
	}

	if err := withdrawA(30050); err != nil { // 300.50 ₽
		fmt.Printf("Снятие %s: %v\n", formatRub(30050), err)
	} else {
		fmt.Printf("Снятие %s -> %s\n", formatRub(30050), formatRub(balanceA()))
	}

	fmt.Printf("Счёт B: %s\n", formatRub(balanceB()))

	// Некорректные суммы: обе операции возвращают одну и ту же ошибку.
	if err := depositA(0); err != nil {
		fmt.Printf("Пополнение %s: %v\n", formatRub(0), err)
	}
	if err := withdrawA(-100); err != nil {
		fmt.Printf("Снятие %s: %v\n", formatRub(-100), err)
	}
}
