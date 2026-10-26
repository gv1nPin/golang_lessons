package main

import (
	"fmt"
	"time"
)

// Hanoi переносит n дисков со стержня from на стержень to через via.
// Что делать с каждым ходом, решает вызывающий код через move.
func Hanoi(n int, from, to, via string, move func(from, to string)) {
	if n == 0 {
		return
	}
	Hanoi(n-1, from, via, to, move)
	move(from, to)
	Hanoi(n-1, via, to, from, move)
}

// numberedPrinter возвращает функцию-ход, которая печатает ходы с номерами.
// Номер хода хранится в замыкании, а не в глобальной переменной.
func numberedPrinter() func(from, to string) {
	step := 0
	return func(from, to string) {
		step++
		fmt.Printf("%d: %s -> %s\n", step, from, to)
	}
}

// countMoves считает ходы для n дисков, не печатая их.
// Время работы выводится через defer после основного вывода.
func countMoves(n int) {
	start := time.Now()
	defer func() {
		fmt.Printf("время: %v\n", time.Since(start))
	}()

	count := 0
	Hanoi(n, "A", "C", "B", func(from, to string) {
		count++
	})
	fmt.Printf("n=%d: %d ходов\n", n, count)
}

// Пусть T(n) это число ходов для n дисков. Чтобы перенести n дисков
// со стержня A на стержень C, нужно:
//  1. перенести n-1 верхних дисков с A на запасной стержень B: T(n-1) ходов;
//  2. переложить самый большой диск с A на C: 1 ход;
//  3. перенести n-1 дисков с B на C: ещё T(n-1) ходов.
//
// Получаем T(n) = 2*T(n-1) + 1, при этом T(0) = 0 (нет дисков, нет ходов).
// Считаем по шагам: T(1) = 1, T(2) = 3, T(3) = 7, T(4) = 15, и так далее,
// то есть T(n) = 2ⁿ − 1. Для n = 20 это 2²⁰ − 1 = 1 048 575 ходов, ровно
// столько и выводит программа.

func main() {
	Hanoi(3, "A", "C", "B", numberedPrinter())
	fmt.Println()
	countMoves(20)
}
