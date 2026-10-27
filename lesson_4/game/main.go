package main

import (
	"io"
)

// TODO: Реализуйте структуры:
// TODO: - Weapon: name (string), weight (float64), Damage (int), Durability (int)
// TODO: - Armor: name (string), weight (float64), Defense (int)
// TODO: - Potion: name (string), weight (float64), Effect (string), Charges (int)
// TODO:
// TODO: name и weight с маленькой буквы: поле Name конфликтовало бы с методом Name()
// TODO:
// TODO: Можете добавить свои структуры и/или немного модифицировать эти :)
// TODO:
// TODO: Для каждой структуры реализуйте:
// TODO: - Метод Use() string (описание использования, например "Используется <имя>" и т.д.)
// TODO: - Методы интерфейса Item

type Item interface {
	Name() string
	Weight() float64
}

// TODO: Добавьте compile-time проверки, что *Weapon, *Armor и *Potion реализуют Item:
// TODO: var _ Item = (*Weapon)(nil)

// TODO: Реализуйте функцию
func DescribeItem(i Item) string {
	// Функция должна возвращать:
	// - "Предмет отсутствует" если i == nil или внутри nil-указатель (var weapon *Weapon)
	// - "<название> (вес: <вес>)" в остальных случаях
	return ""
}

func Filter[T any](items []T, predicate func(T) bool) []T {
	// TODO: Верните новый слайс только с элементами, для которых predicate вернул true
	return nil
}

func Map[T any, R any](items []T, transform func(T) R) []R {
	// TODO: Примените transform к каждому элементу и верните слайс с результатами
	return nil
}

func Find[T any](items []T, condition func(T) bool) (T, bool) {
	// TODO: Найдите первый элемент, удовлетворяющий condition и верните элемент и true или zero value и false
	var zero T
	return zero, false
}

type Inventory struct {
	Items []Item
}

func (inv *Inventory) AddItem(item Item) {
	inv.Items = append(inv.Items, item)
}

func (inv *Inventory) Weapons() []*Weapon {
	// TODO: Используйте Filter для отбора Weapon, затем Map для преобразования Item -> *Weapon
	return nil
}

func (inv *Inventory) BrokenItems() []Item {
	// TODO: Используйте Filter для отбора:
	// TODO: - Weapon: Durability <= 0
	// TODO: - Potion: Charges <= 0
	// TODO:
	// TODO: Подсказка: поможет приведение типов - item.(type)
	return nil
}

func (inv *Inventory) ItemNames() []string {
	// TODO: Используйте Map для преобразования []Item -> []string
	return nil
}

func (inv *Inventory) FindItemByName(name string) (Item, bool) {
	// TODO: Используйте Find для поиска по имени
	return nil, false
}

// TODO: Бонус: реализуйте интефейс Storable для Weapon и Armor:
// TODO: - Weapon: формат "Weapon|Name|Damage|Durability|Weight"
// TODO: - Armor: формат "Armor|Name|Defense|Weight"

type Storable interface {
	Serialize(w io.Writer)
	Deserialize(r io.Reader)
}

func (inv *Inventory) Save(w io.Writer) {
	// TODO: Бонус: сделайте сохранение/загрузку инвентаря в/из файла
}

func (inv *Inventory) Load(r io.Reader) {
	// TODO: Бонус: сделайте сохранение/загрузку инвентаря в/из файла
}

func main() {
	// TODO: Создайте инвентарь и добавьте:
	// TODO: - Оружие: "Меч" (урон 10, прочность 5)
	// TODO: - Броню: "Щит" (защита 5, вес 4.5)
	// TODO: - Зелье: "Лечебное" (+50 HP, 3 заряда)
	// TODO: - Оружие: "Сломанный лук" (урон 5, прочность 0)

	// TODO: Реализуйте логику/вызовы:
	// TODO: 1. Use предмета с выводом в консоль
	// TODO: 2. DescribeItem с предметом, с nil и с nil-указателем (var weapon *Weapon), так же с выводом в консоль
	// TODO: 3. Вывести в консоль результат вызова Weapons (должны вернуться только меч и лук)
	// TODO: 4. Вывести в консоль результат вызова BrokenItems (должен вернуть сломанный лук)
	// TODO: 5. Вывести в консоль результат вызова ItemNames (все названия)
	// TODO: 6. Вывести в консоль результат вызова FindItemByName (поиск "Щит")

	// TODO: Бонус: сделайте сохранение инвентаря в файл и загрузку инвентаря из файла
}
