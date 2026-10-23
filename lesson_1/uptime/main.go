package main

import (
	"fmt"
	"strings"
)

const (
	secondsInMinute = 60
	secondsInHour   = 60 * secondsInMinute
	secondsInDay    = 24 * secondsInHour
)

func formatUptime(total int) string {
	days := total / secondsInDay
	remainder := total % secondsInDay

	hours := remainder / secondsInHour
	remainder %= secondsInHour

	minutes := remainder / secondsInMinute
	seconds := remainder % secondsInMinute

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dд", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dч", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dм", minutes))
	}
	if seconds > 0 {
		parts = append(parts, fmt.Sprintf("%dс", seconds))
	}

	if len(parts) == 0 {
		return "0с"
	}
	return strings.Join(parts, " ")
}

func main() {
	examples := []int{93784, 86405, 3600, 59, 0}

	for _, total := range examples {
		fmt.Printf("%d -> %s\n", total, formatUptime(total))
	}
}
