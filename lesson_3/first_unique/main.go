package main

import "fmt"

func FirstUnique(s string) rune {
	counts := make(map[rune]int)
	for _, r := range s {
		counts[r]++
	}

	for _, r := range s {
		if counts[r] == 1 {
			return r
		}
	}
	return 0
}

func main() {
	inputs := []string{
		"abacabad",
		"abcabcd",
		"абракадабра",
		"aabbcc",
	}

	for _, s := range inputs {
		fmt.Printf("FirstUnique(%q) -> %q\n", s, FirstUnique(s))
	}
}
