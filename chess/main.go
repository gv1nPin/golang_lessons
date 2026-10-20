package main

import "fmt"

func main() {
	rowEven := " # # # #"
	rowOdd := "# # # # "

	const N = 8

	for row := 0; row < N; row++ {
		if row%2 == 0 {
			fmt.Println(rowEven)
		} else {
			fmt.Println(rowOdd)
		}
	}
}
