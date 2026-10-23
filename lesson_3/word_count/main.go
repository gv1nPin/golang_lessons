package main

import (
	"fmt"
	"strings"
)

func WordCount(s string) map[string]int {
	counts := make(map[string]int)
	for _, word := range strings.Fields(s) {
		counts[word]++
	}
	return counts
}

func main() {
	inputs := []string{
		"go is fun go",
		"Hello hello World",
		" мир мир труд ",
	}

	for _, s := range inputs {
		fmt.Printf("WordCount(%q) -> %v\n", s, WordCount(s))
	}
}
