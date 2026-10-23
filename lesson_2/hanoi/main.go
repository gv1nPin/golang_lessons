package main

import (
	"fmt"
	"time"
)

func Hanoi(n int, from, to, via string, move func(from, to string)) {
	if n == 0 {
		return
	}
	Hanoi(n-1, from, via, to, move)
	move(from, to)
	Hanoi(n-1, via, to, from, move)
}

func numberedPrinter() func(from, to string) {
	step := 0
	return func(from, to string) {
		step++
		fmt.Printf("%d: %s -> %s\n", step, from, to)
	}
}

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

func main() {
	Hanoi(3, "A", "C", "B", numberedPrinter())
	fmt.Println()
	countMoves(20)
}
