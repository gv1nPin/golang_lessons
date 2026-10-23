package main

import "fmt"

func main() {
	fmt.Println("Таблица температур")
	fmt.Println("------------------------")

	for celsius := -20; celsius <= 40; celsius += 10 {
		fahrenheit := float64(celsius)*9.0/5.0 + 32.0

	}
}
