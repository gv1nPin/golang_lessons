package main

import "fmt"

// BracketsOK проверяет, что скобки (), [], {} в строке расставлены верно:
// каждая закрывающая скобка закрывает ближайшую незакрытую открывающую
// скобку того же вида, и в конце незакрытых скобок не остаётся.
// Любые другие символы игнорируются.
//
// Почему не хватает счётчика: для «([)]» он скажет «верно» (1, 2, 1, 0),
// хотя ) закрывает [. Счётчик не помнит, КАКИЕ скобки открыты и в каком
// порядке. Это помнит стек.
//
// Шаблон «Стек»:
//  1. пустой стек;
//  2. открывающая скобка -> положить на стек;
//  3. закрывающая скобка -> стек пуст? верх подходит? снять;
//  4. в конце стек должен быть пуст.
//
// Время: O(n), память: O(n).
func BracketsOK(s string) bool {
	pair := map[rune]rune{
		')': '(',
		']': '[',
		'}': '{',
	}

	stack := []rune{} // 1

	for _, ch := range s {
		switch ch {
		case '(', '[', '{': // 2
			stack = append(stack, ch)
		case ')', ']', '}': // 3
			if len(stack) == 0 {
				return false // снимать нечего: в Go это была бы panic
			}
			top := stack[len(stack)-1]
			if top != pair[ch] {
				return false
			}
			stack = stack[:len(stack)-1]
		}
	}

	return len(stack) == 0 // 4
}

func main() {
	fmt.Println(BracketsOK("([{}])")) // true
	fmt.Println(BracketsOK("([)]"))   // false
	fmt.Println(BracketsOK(")("))     // false
	fmt.Println(BracketsOK("(("))     // false
}
