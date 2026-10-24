package main

import "fmt"

func CountBits(n uint8) int {
	count := 0
	for n > 0 {
		count += int(n & 1)
		n >>= 1
	}
	return count
}

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
