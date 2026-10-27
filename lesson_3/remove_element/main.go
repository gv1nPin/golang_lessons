package main

import "fmt"

// RemoveElement возвращает новый слайс без элемента с индексом index,
// порядок остальных элементов сохраняется. Исходный слайс не меняется.
// При недопустимом индексе (в том числе отрицательном) возвращает nil и ошибку.
func RemoveElement(nums []int, index int) ([]int, error) {
	if index < 0 || index >= len(nums) {
		return nil, fmt.Errorf("индекс %d вне диапазона [0, %d)", index, len(nums))
	}

	// Нельзя писать append(nums[:index], nums[index+1:]...): это сдвинет
	// элементы внутри исходного массива и изменит nums. Собираем копию.
	result := make([]int, 0, len(nums)-1)
	result = append(result, nums[:index]...)
	result = append(result, nums[index+1:]...)
	return result, nil
}

// Слайс не хранит данные сам, он указывает на массив. Слайсы nums и
// nums[:index] указывают на один и тот же массив, и append при достаточной
// ёмкости пишет прямо в него: сдвигает элементы после index на одну позицию
// влево. Исходный слайс при этом портится. Для [1 2 2 3 1] и index = 3
// получилось бы [1 2 2 1 1]: длина прежняя, а последний элемент продублирован.
// Поэтому результат собираем в новом слайсе (make и два append), а массив
// исходного слайса остаётся нетронутым.

func main() {
	original := []int{1, 2, 2, 3, 1}
	result, err := RemoveElement(original, 3)
	fmt.Printf("RemoveElement(%v, 3) -> %v %v, исходный слайс: %v\n", original, result, err, original)

	short := []int{1, 2, 3}
	for _, index := range []int{5, -1} {
		result, err := RemoveElement(short, index)
		fmt.Printf("RemoveElement(%v, %d) -> %v %v\n", short, index, result, err)
	}
}
