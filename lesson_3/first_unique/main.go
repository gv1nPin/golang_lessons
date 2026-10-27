package main

import "fmt"

// FirstUnique возвращает первый символ, который встречается в строке
// ровно один раз. Если такого нет, возвращает 0. Поиск регистрозависимый.
// Идём по рунам, а не по байтам, чтобы кириллица работала правильно.
func FirstUnique(s string) rune {
	counts := make(map[rune]int)
	for _, r := range s {
		counts[r]++
	}

	// map не хранит порядок, поэтому второй проход идёт по самой строке.
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
