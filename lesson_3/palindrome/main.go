package main

import (
	"fmt"
	"unicode"
)

// IsPalindrome проверяет, читается ли строка одинаково в обе стороны.
// Регистр и все небуквенные символы (пробелы, знаки препинания, цифры)
// игнорируются. Строка без букв считается палиндромом.
func IsPalindrome(s string) bool {
	var letters []rune
	for _, r := range s {
		if unicode.IsLetter(r) {
			letters = append(letters, unicode.ToLower(r))
		}
	}

	for i, j := 0, len(letters)-1; i < j; i, j = i+1, j-1 {
		if letters[i] != letters[j] {
			return false
		}
	}
	return true
}

func main() {
	inputs := []string{
		"шалаш",
		"А роза упала на лапу Азора",
		"Was it a car or a cat I saw?",
		"hello",
	}

	for _, s := range inputs {
		call := fmt.Sprintf("IsPalindrome(%q)", s)
		fmt.Printf("%-46s -> %t\n", call, IsPalindrome(s))
	}
}
