package main

import "testing"

func TestHasDuplicate(t *testing.T) {
	cases := []struct {
		name string
		nums []int
		want bool
	}{
		{"nil-срез", nil, false},
		{"пустой срез", []int{}, false},
		{"один элемент", []int{7}, false},
		{"все разные", []int{1, 2, 3, 4}, false},
		{"дубликат рядом", []int{1, 2, 2, 3}, true},
		{"дубликат в начале и в конце", []int{5, 1, 2, 3, 5}, true},
		{"все одинаковые", []int{4, 4, 4, 4}, true},
		{"отрицательные числа и ноль", []int{-1, 0, 1, -1}, true},
		{"отрицательные без повторов", []int{-3, -2, -1, 0}, false},
	}

	for _, c := range cases {
		got := HasDuplicate(c.nums)
		if got != c.want {
			t.Errorf("%s: получили %v, ожидали %v", c.name, got, c.want)
		}
	}
}
