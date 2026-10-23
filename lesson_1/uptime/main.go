package main

import "fmt"

func main() {
	totalSeconds := 93784

	const (
		secondsIsMinute = 60
		secondsIsHour   = 60 * 60
		secondsIsDay    = 60 * 60 * 24
	)

	days := totalSeconds / secondsIsDay
	remainder := totalSeconds % secondsIsDay

	hours := remainder / secondsIsHour
	remainder = remainder % secondsIsHour

	minutes := remainder / secondsIsMinute
	seconds := remainder % secondsIsMinute

	hasOutput := false

	if days > 0 {
		fmt.Printf("%d days\n", days)
		hasOutput = true
	}
	if hours > 0 {
		fmt.Printf("%d hours\n", hours)
		hasOutput = true
	}
	if minutes > 0 {
		fmt.Printf("%d minutes\n", minutes)
		hasOutput = true
	}
	if seconds > 0 {
		fmt.Printf("%d seconds\n", seconds)
		hasOutput = true
	}
	if !hasOutput {
		fmt.Printf("0c")
	}
	fmt.Printf("\n")
}
