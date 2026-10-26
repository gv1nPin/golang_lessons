package main

import "fmt"

// CountBits считает единичные биты числа через & и сдвиг >>.
func CountBits(n uint8) int {
	count := 0
	for n > 0 {
		count += int(n & 1)
		n >>= 1
	}
	return count
}

// IsPowerOfTwo проверяет степень двойки одним выражением.
// У степени двойки ровно один единичный бит, а у n-1 все младшие биты
// от него единичные: 8 = 00001000, 7 = 00000111, их & равен нулю.
// Для 6 = 00000110 и 5 = 00000101 результат 00000100, то есть не ноль.
// Отдельно проверяем n != 0, потому что 0 & (0-1) тоже даёт 0.
func IsPowerOfTwo(n uint) bool {
	return n != 0 && n&(n-1) == 0
}

func main() {
	fmt.Println("CountBits:")
	for _, n := range []uint8{0, 5, 12, 255} {
		fmt.Printf("%d = %08b -> %d\n", n, n, CountBits(n))
	}

	fmt.Println()
	fmt.Println("IsPowerOfTwo:")
	for _, n := range []uint{1, 6, 64, 0} {
		fmt.Printf("%d -> %t\n", n, IsPowerOfTwo(n))
	}

	fmt.Println()
	fmt.Println("n, n-1 и n & (n-1):")
	for _, n := range []uint{8, 6} {
		fmt.Printf("n = %08b, n-1 = %08b, n&(n-1) = %08b\n", n, n-1, n&(n-1))
	}
}
