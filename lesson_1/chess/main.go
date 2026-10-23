package main

import (
	"fmt"
	"strings"
)

func main() {
	const N = 8

	for row := 0; row < N; row++ {
		var line strings.Builder

		for col := 0; col < N; col++ {
			if (row+col)%2 == 0 {
				line.WriteString(" ")
			} else {
				line.WriteString("#")
			}
		}
		fmt.Println(line.String())
	}
}
