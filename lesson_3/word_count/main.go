package main

import (
	"fmt"
	"strings"
)

// WordCount возвращает map: слово -> число вхождений.
// Слова разделяются пробелами (их может быть несколько подряд),
// регистр учитывается: Go и go считаются разными словами.
// strings.Fields делит строку по любому количеству пробелов подряд и не
// создаёт пустых слов, поэтому "go    is" даёт два слова, а не пять.
// Для пустой строки возвращается пустая (не nil) map.
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
		"go    is   fun    go",
		"   Hello    hello  World   ",
	}

	// fmt печатает ключи map отсортированными: это свойство вывода,
	// а не порядок обхода map через range.
	for _, s := range inputs {
		fmt.Printf("WordCount(%q) -> %v\n", s, WordCount(s))
	}
}
