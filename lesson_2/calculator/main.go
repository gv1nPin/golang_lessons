package main

import (
	"errors"
	"fmt"
)

type Operation func(a, b float64) (float64, error)

var ErrDivisionByZero = errors.New("деление на ноль")

var operations = map[string]Operation{
	"+": func(a, b float64) (float64, error) {
		return a + b, nil
	},
	"-": func(a, b float64) (float64, error) {
		return a - b, nil
	},
	"*": func(a, b float64) (float64, error) {
		return a * b, nil
	},
	"/": func(a, b float64) (float64, error) {
		if b == 0 {
			return 0, ErrDivisionByZero
		}
		return a / b, nil
	},
}

// Calculate ищет операцию в map и вызывает её.
// Для неизвестной операции возвращает ошибку и ничего не вызывает.
func Calculate(a float64, op string, b float64) (float64, error) {
	operation, ok := operations[op]
	if !ok {
		return 0, fmt.Errorf("неизвестная операция %q", op)
	}
	return operation(a, b)
}

func main() {
	cases := []struct {
		a  float64
		op string
		b  float64
	}{
		{10, "+", 5},
		{10, "/", 4},
		{7, "*", 6},
		{1, "/", 0},
		{2, "%", 3},
	}

	for _, c := range cases {
		result, err := Calculate(c.a, c.op, c.b)
		if err != nil {
			fmt.Printf("%v %s %v: %v\n", c.a, c.op, c.b, err)
			continue
		}
		fmt.Printf("%v %s %v = %.2f\n", c.a, c.op, c.b, result)
	}
}
