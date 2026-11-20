package main

import "fmt"

// HasDuplicate возвращает true, если в срезе есть хотя бы два одинаковых числа.
//
// Шаблон «Видел ли я уже?»:
//  1. создаём пустое множество seen (в Go это map[int]bool);
//  2. идём по элементам;
//  3. если число уже есть в seen - дубликат найден;
//  4. иначе добавляем число в seen.
//
// Время: O(n), память: O(n).
// Наивное решение «каждый с каждым» работало бы за O(n²) при памяти O(1).
// Входной срез не изменяется.
func HasDuplicate(nums []int) bool {
	seen := make(map[int]bool) // 1

	for _, x := range nums { // 2
		if seen[x] { // 3
			return true
		}
		seen[x] = true // 4
	}

	return false
}

func main() {
	fmt.Println(HasDuplicate([]int{1, 2, 3, 1})) // true
	fmt.Println(HasDuplicate([]int{1, 2, 3, 4})) // false
}
