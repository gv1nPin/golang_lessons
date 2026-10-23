package main

import "fmt"

func main() {
	fmt.Println("Таблица температур")
	fmt.Println("------------------------")

	for celsius := -20; celsius <= 40; celsius += 10 {
		fahrenheit := float64(celsius)*9.0/5.0 + 32.0

		var description string
		switch {
		case celsius >= 25:
			description = "жарко"
		case celsius >= 15:
			description = "тепло"
		case celsius >= 0:
			description = "холодно"
		default:
			description = "мороз"

		}
		fmt.Printf("%3d°C = %5.1f°F  %s\n", celsius, fahrenheit, description)
	}
}
