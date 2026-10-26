package main

import "fmt"

func main() {
	for celsius := -20; celsius <= 40; celsius += 10 {
		fahrenheit := float64(celsius)*9.0/5.0 + 32.0

		var description string
		switch {
		case celsius < 0:
			description = "мороз"
		case celsius < 15:
			description = "холодно"
		case celsius < 25:
			description = "тепло"
		default:
			description = "жарко"
		}

		fmt.Printf("%3d°C = %5.1f°F  %s\n", celsius, fahrenheit, description)
	}
}
