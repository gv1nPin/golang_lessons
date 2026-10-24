package main

import (
	"fmt"
	"unicode"
)

func letterCounts(s string) map[rune]int {
	counts := make(map[rune]int)
	for _, r := range s {
		if unicode.IsSpace(r) {
			continue
		}
		counts[unicode.ToLower(r)]++
	}
	return counts
}

func AreAnagrams(s1, s2 string) bool {
	a, b := letterCounts(s1), letterCounts(s2)
	if len(a) != len(b) {
		return false
	}
	for r, n := range a {
		if b[r] != n {
			return false
		}
	}
	return true
}

func main() {
	pairs := [][2]string{
		{"listen", "silent"},
		{"Eleven plus two", "Twelve plus one"},
		{"Апельсин", "спаниель"},
		{"Hello", "World"},
	}

	for _, p := range pairs {
		fmt.Printf("AreAnagrams(%q, %q) -> %t\n", p[0], p[1], AreAnagrams(p[0], p[1]))
	}
}
