package main

import "testing"

func TestTwoItems(t *testing.T) {
	cases := []struct {
		name   string
		prices []int
		k      int
		i, j   int
		wantOK bool
	}{
		{"пара в середине",
			[]int{700, 1500, 300, 1200}, 1500, 2, 3, true},
		{"ответ на позиции 0",
			[]int{700, 800}, 1500, 0, 1, true},
		{"нельзя взять один товар дважды",
			[]int{700, 500}, 1400, 0, 0, false},
		{"две одинаковые цены - это два разных товара",
			[]int{700, 700}, 1400, 0, 1, true},
		{"нет пары",
			[]int{1, 2, 3}, 100, 0, 0, false},
		{"пустой список",
			[]int{}, 10, 0, 0, false},
		{"nil-список",
			nil, 10, 0, 0, false},
		{"один товар",
			[]int{10}, 10, 0, 0, false},
		{"нулевая цена",
			[]int{0, 5}, 5, 0, 1, true},
		{"пара в самом конце",
			[]int{1, 2, 3, 4, 100, 200}, 300, 4, 5, true},
	}

	for _, c := range cases {
		i, j, ok := TwoItems(c.prices, c.k)
		if ok != c.wantOK || (ok && (i != c.i || j != c.j)) {
			t.Errorf("%s: получили %d, %d, %v", c.name, i, j, ok)
		}
	}
}
