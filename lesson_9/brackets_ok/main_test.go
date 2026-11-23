package main

import "testing"

func TestBracketsOK(t *testing.T) {
	cases := []struct {
		name string
		s    string
		want bool
	}{
		{"пустая строка", "", true},
		{"одна пара", "()", true},
		{"три вида подряд", "()[]{}", true},
		{"вложенные", "([{}])", true},
		{"неверный порядок закрытия", "([)]", false},
		{"закрывающая раньше открывающей", ")(", false},
		{"не закрыты", "((", false},
		{"одна открывающая", "(", false},
		{"одна закрывающая", ")", false},
		{"разные виды в паре", "(]", false},
		{"лишняя закрывающая в конце", "()]", false},
		{"лишняя открывающая в начале", "{()", false},
	}

	for _, c := range cases {
		got := BracketsOK(c.s)
		if got != c.want {
			t.Errorf("%s: BracketsOK(%q) = %v, ожидали %v", c.name, c.s, got, c.want)
		}
	}
}
