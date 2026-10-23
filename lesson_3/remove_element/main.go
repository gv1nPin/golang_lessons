package main

import "fmt"

func RemoveElement(nums []int, index int) ([]int, error) {
	if index < 0 || index >= len(nums) {
		return nil, fmt.Errorf("индекс %d вне диапазона [0, %d)", index, len(nums))
	}

	result := make([]int, 0, len(nums)-1)
	result = append(result, nums[:index]...)
	result = append(result, nums[index+1:]...)
	return result, nil
}

func main() {
	original := []int{1, 2, 2, 3, 1}
	result, err := RemoveElement(original, 3)
	fmt.Printf("RemoveElement(%v, 3) -> %v %v, исходный слайс: %v\n", []int{1, 2, 2, 3, 1}, result, err, original)

	short := []int{1, 2, 3}
	for _, index := range []int{5, -1} {
		result, err := RemoveElement(short, index)
		fmt.Printf("RemoveElement(%v, %d) -> %v %v\n", short, index, result, err)
	}
}
