package main

import "fmt"

func main() {
	rowEven := " # # # #"
	rowOdd := "# # # # "

	for row := 0; row < 8; row++ {
		if row%2 == 0 {
			fmt.Println(rowEven)
		} else {
			fmt.Println(rowOdd)
		}
	}
}
