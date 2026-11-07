package inventory

// Обобщённые помощники нужны только методам Inventory, поэтому не экспортируются.

// filter возвращает новый слайс только с теми элементами, для которых predicate вернул true.
func filter[T any](items []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(items))
	for _, item := range items {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// mapSlice применяет transform к каждому элементу и возвращает слайс с результатами.
func mapSlice[T any, R any](items []T, transform func(T) R) []R {
	result := make([]R, 0, len(items))
	for _, item := range items {
		result = append(result, transform(item))
	}
	return result
}

// find возвращает первый элемент, для которого condition вернул true, и true.
// Если такого элемента нет, возвращает нулевое значение и false.
func find[T any](items []T, condition func(T) bool) (T, bool) {
	for _, item := range items {
		if condition(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}
