package inventory

import (
	"slices"
	"strconv"
	"testing"
)

func TestFilter(t *testing.T) {
	isEven := func(n int) bool { return n%2 == 0 }

	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"nil", nil, nil},
		{"пустой слайс", []int{}, nil},
		{"подходят все", []int{2, 4, 6}, []int{2, 4, 6}},
		{"не подходит никто", []int{1, 3, 5}, nil},
		{"подходят некоторые, порядок сохраняется", []int{1, 2, 3, 4, 5, 6}, []int{2, 4, 6}},
		{"один элемент", []int{2}, []int{2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filter(tt.in, isEven); !slices.Equal(got, tt.want) {
				t.Errorf("filter(%v) = %v, хочу %v", tt.in, got, tt.want)
			}
		})
	}
}

// filter возвращает новый слайс: запись в результат не меняет исходные данные.
func TestFilterReturnsNewSlice(t *testing.T) {
	in := []int{1, 2, 3, 4}

	out := filter(in, func(n int) bool { return n > 0 })
	out[0] = 100

	if want := []int{1, 2, 3, 4}; !slices.Equal(in, want) {
		t.Errorf("исходный слайс изменился: %v, хочу %v", in, want)
	}
}

func TestMapSlice(t *testing.T) {
	t.Run("числа в строки", func(t *testing.T) {
		tests := []struct {
			name string
			in   []int
			want []string
		}{
			{"nil", nil, nil},
			{"пустой слайс", []int{}, nil},
			{"порядок сохраняется", []int{3, 1, 2}, []string{"3", "1", "2"}},
			{"отрицательные", []int{-1, 0}, []string{"-1", "0"}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := mapSlice(tt.in, strconv.Itoa); !slices.Equal(got, tt.want) {
					t.Errorf("mapSlice(%v) = %v, хочу %v", tt.in, got, tt.want)
				}
			})
		}
	})

	t.Run("тот же тип", func(t *testing.T) {
		got := mapSlice([]int{1, 2, 3}, func(n int) int { return n * n })
		if want := []int{1, 4, 9}; !slices.Equal(got, want) {
			t.Errorf("mapSlice = %v, хочу %v", got, want)
		}
	})

	t.Run("не меняет исходный слайс", func(t *testing.T) {
		in := []int{1, 2, 3}
		mapSlice(in, func(n int) int { return n + 10 })
		if want := []int{1, 2, 3}; !slices.Equal(in, want) {
			t.Errorf("исходный слайс изменился: %v", in)
		}
	})
}

func TestFind(t *testing.T) {
	isBig := func(n int) bool { return n > 10 }

	tests := []struct {
		name      string
		in        []int
		want      int
		wantFound bool
	}{
		{"nil", nil, 0, false},
		{"пустой слайс", []int{}, 0, false},
		{"не найден", []int{1, 2, 3}, 0, false},
		{"найден единственный", []int{1, 20, 3}, 20, true},
		{"возвращается первый из подходящих", []int{1, 20, 30}, 20, true},
		{"подходит первый элемент", []int{11, 2}, 11, true},
		{"подходит последний элемент", []int{1, 2, 12}, 12, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := find(tt.in, isBig)
			if found != tt.wantFound || got != tt.want {
				t.Errorf("find(%v) = %v, %v; хочу %v, %v", tt.in, got, found, tt.want, tt.wantFound)
			}
		})
	}
}

// Если ничего не найдено, find возвращает нулевое значение своего типа.
func TestFindZeroValue(t *testing.T) {
	type pair struct{ a, b int }

	if got, found := find([]string{"a"}, func(s string) bool { return s == "z" }); found || got != "" {
		t.Errorf("строки: %q, %v", got, found)
	}
	if got, found := find([]pair{{1, 2}}, func(p pair) bool { return p.a == 9 }); found || got != (pair{}) {
		t.Errorf("структуры: %v, %v", got, found)
	}
	if got, found := find([]*Weapon{}, func(w *Weapon) bool { return true }); found || got != nil {
		t.Errorf("указатели: %v, %v", got, found)
	}
}
