package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// name и weight с маленькой буквы: поле Name конфликтовало бы с методом Name()

type Weapon struct {
	name       string
	weight     float64
	Damage     int
	Durability int
}

type Armor struct {
	name    string
	weight  float64
	Defense int
}

type Potion struct {
	name    string
	weight  float64
	Effect  string
	Charges int
}

type Item interface {
	Name() string
	Weight() float64
}

// Compile-time проверки: если метод забыт или имеет не ту сигнатуру,
// код не скомпилируется.
var _ Item = (*Weapon)(nil)
var _ Item = (*Armor)(nil)
var _ Item = (*Potion)(nil)

// --- Weapon ---

func (w *Weapon) Name() string    { return w.name }
func (w *Weapon) Weight() float64 { return w.weight }

func (w *Weapon) Use() string {
	if w.Durability <= 0 {
		return fmt.Sprintf("%s сломано и непригодно для боя", w.name)
	}
	return fmt.Sprintf("Используется %s: урон %d (прочность: %d)", w.name, w.Damage, w.Durability)
}

// --- Armor ---

func (a *Armor) Name() string    { return a.name }
func (a *Armor) Weight() float64 { return a.weight }

func (a *Armor) Use() string {
	return fmt.Sprintf("Используется %s: защита +%d", a.name, a.Defense)
}

// --- Potion ---

func (p *Potion) Name() string    { return p.name }
func (p *Potion) Weight() float64 { return p.weight }

func (p *Potion) Use() string {
	if p.Charges <= 0 {
		return fmt.Sprintf("%s закончилось", p.name)
	}
	return fmt.Sprintf("Используется %s: %s (осталось зарядов: %d)", p.name, p.Effect, p.Charges)
}

// isNilItem сообщает, что предмета фактически нет. Таких случаев два:
//  1. в интерфейсе вообще ничего нет (i == nil);
//  2. интерфейс не пустой, но внутри лежит nil-указатель (var weapon *Weapon).
//     Тогда i != nil, а вызов i.Name() упал бы с паникой.
func isNilItem(i Item) bool {
	switch v := i.(type) {
	case nil:
		return true
	case *Weapon:
		return v == nil
	case *Armor:
		return v == nil
	case *Potion:
		return v == nil
	}
	return false
}

func DescribeItem(i Item) string {
	if isNilItem(i) {
		return "Предмет отсутствует"
	}
	return fmt.Sprintf("%s (вес: %.1f)", i.Name(), i.Weight())
}

func Filter[T any](items []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(items))
	for _, item := range items {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

func Map[T any, R any](items []T, transform func(T) R) []R {
	result := make([]R, 0, len(items))
	for _, item := range items {
		result = append(result, transform(item))
	}
	return result
}

func Find[T any](items []T, condition func(T) bool) (T, bool) {
	for _, item := range items {
		if condition(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

type Inventory struct {
	Items []Item
}

// AddItem не принимает пустые предметы: ни nil, ни nil-указатель.
// Иначе методы вроде ItemNames или Save упали бы с паникой.
func (inv *Inventory) AddItem(item Item) error {
	if isNilItem(item) {
		return errors.New("нельзя добавить пустой предмет")
	}
	inv.Items = append(inv.Items, item)
	return nil
}

func (inv *Inventory) Weapons() []*Weapon {
	weaponItems := Filter(inv.Items, func(item Item) bool {
		_, ok := item.(*Weapon)
		return ok
	})
	return Map(weaponItems, func(item Item) *Weapon {
		return item.(*Weapon)
	})
}

func (inv *Inventory) BrokenItems() []Item {
	return Filter(inv.Items, func(item Item) bool {
		switch v := item.(type) {
		case *Weapon:
			return v.Durability <= 0
		case *Potion:
			return v.Charges <= 0
		}
		return false
	})
}

func (inv *Inventory) ItemNames() []string {
	return Map(inv.Items, func(item Item) string {
		return item.Name()
	})
}

func (inv *Inventory) FindItemByName(name string) (Item, bool) {
	return Find(inv.Items, func(item Item) bool {
		return item.Name() == name
	})
}

// --- Бонус: сохранение и загрузка ---
//
// Методы возвращают error: запись и разбор могут завершиться ошибкой,
// и молча её терять нельзя.

type Storable interface {
	Serialize(w io.Writer) error
	Deserialize(r io.Reader) error
}

// StorableItem — предмет, который можно сохранить.
type StorableItem interface {
	Item
	Storable
}

var _ Storable = (*Weapon)(nil)
var _ Storable = (*Armor)(nil)
var _ Storable = (*Potion)(nil) // Potion добавлен сверх ТЗ, чтобы зелья не терялись при сохранении

// checkField отклоняет значения, которые сломали бы формат записи:
// разделитель "|" и перевод строки (одна запись = одна строка файла).
func checkField(label, value string) error {
	if strings.ContainsAny(value, "|\r\n") {
		return fmt.Errorf("%s %q содержит символ «|» или перевод строки", label, value)
	}
	return nil
}

// Формат: "Weapon|Name|Damage|Durability|Weight"
func (w *Weapon) Serialize(out io.Writer) error {
	if err := checkField("название оружия", w.name); err != nil {
		return err
	}
	_, err := fmt.Fprintf(out, "Weapon|%s|%d|%d|%v\n", w.name, w.Damage, w.Durability, w.weight)
	return err
}

func (w *Weapon) Deserialize(r io.Reader) error {
	parts, err := readRecord(r, "Weapon", 5)
	if err != nil {
		return err
	}
	damage, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("неверный урон %q: %w", parts[2], err)
	}
	durability, err := strconv.Atoi(parts[3])
	if err != nil {
		return fmt.Errorf("неверная прочность %q: %w", parts[3], err)
	}
	weight, err := strconv.ParseFloat(parts[4], 64)
	if err != nil {
		return fmt.Errorf("неверный вес %q: %w", parts[4], err)
	}

	w.name = parts[1]
	w.Damage = damage
	w.Durability = durability
	w.weight = weight
	return nil
}

// Формат: "Armor|Name|Defense|Weight"
func (a *Armor) Serialize(out io.Writer) error {
	if err := checkField("название брони", a.name); err != nil {
		return err
	}
	_, err := fmt.Fprintf(out, "Armor|%s|%d|%v\n", a.name, a.Defense, a.weight)
	return err
}

func (a *Armor) Deserialize(r io.Reader) error {
	parts, err := readRecord(r, "Armor", 4)
	if err != nil {
		return err
	}
	defense, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("неверная защита %q: %w", parts[2], err)
	}
	weight, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return fmt.Errorf("неверный вес %q: %w", parts[3], err)
	}

	a.name = parts[1]
	a.Defense = defense
	a.weight = weight
	return nil
}

// Формат: "Potion|Name|Effect|Charges|Weight"
func (p *Potion) Serialize(out io.Writer) error {
	if err := checkField("название зелья", p.name); err != nil {
		return err
	}
	if err := checkField("эффект зелья", p.Effect); err != nil {
		return err
	}
	_, err := fmt.Fprintf(out, "Potion|%s|%s|%d|%v\n", p.name, p.Effect, p.Charges, p.weight)
	return err
}

func (p *Potion) Deserialize(r io.Reader) error {
	parts, err := readRecord(r, "Potion", 5)
	if err != nil {
		return err
	}
	charges, err := strconv.Atoi(parts[3])
	if err != nil {
		return fmt.Errorf("неверное число зарядов %q: %w", parts[3], err)
	}
	weight, err := strconv.ParseFloat(parts[4], 64)
	if err != nil {
		return fmt.Errorf("неверный вес %q: %w", parts[4], err)
	}

	p.name = parts[1]
	p.Effect = parts[2]
	p.Charges = charges
	p.weight = weight
	return nil
}

// readRecord читает одну запись, делит её по "|" и проверяет тип и число полей.
func readRecord(r io.Reader, kind string, fields int) ([]string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	record := strings.TrimSpace(string(data))
	parts := strings.Split(record, "|")
	if len(parts) != fields || parts[0] != kind {
		return nil, fmt.Errorf("неверный формат записи %s: %q", kind, record)
	}
	return parts, nil
}

// Save записывает по одной строке на предмет.
// Предметы, которые не реализуют Storable, пропускаются.
func (inv *Inventory) Save(w io.Writer) error {
	for _, item := range inv.Items {
		storable, ok := item.(Storable)
		if !ok {
			continue
		}
		if err := storable.Serialize(w); err != nil {
			return err
		}
	}
	return nil
}

// Load заменяет содержимое инвентаря предметами из r.
// Если в данных есть ошибка, инвентарь остаётся без изменений.
func (inv *Inventory) Load(r io.Reader) error {
	var items []Item

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		kind, _, _ := strings.Cut(line, "|")

		var item StorableItem
		switch kind {
		case "Weapon":
			item = &Weapon{}
		case "Armor":
			item = &Armor{}
		case "Potion":
			item = &Potion{}
		default:
			return fmt.Errorf("неизвестный тип предмета: %q", kind)
		}

		if err := item.Deserialize(strings.NewReader(line)); err != nil {
			return err
		}
		items = append(items, item)
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	inv.Items = items
	return nil
}

func saveToFile(inv *Inventory, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := inv.Save(file); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func loadFromFile(inv *Inventory, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return inv.Load(file)
}

func main() {
	sword := &Weapon{name: "Меч", weight: 3.5, Damage: 10, Durability: 5}
	shield := &Armor{name: "Щит", weight: 4.5, Defense: 5}
	potion := &Potion{name: "Лечебное", weight: 0.3, Effect: "+50 HP", Charges: 3}
	bow := &Weapon{name: "Сломанный лук", weight: 1.2, Damage: 5, Durability: 0}

	inv := &Inventory{}
	for _, item := range []Item{sword, shield, potion, bow} {
		if err := inv.AddItem(item); err != nil {
			fmt.Println("Ошибка добавления:", err)
			return
		}
	}

	// AddItem: пустые предметы отклоняются
	fmt.Println("--- AddItem ---")
	var weapon *Weapon // nil-указатель
	fmt.Println("nil:", inv.AddItem(nil))
	fmt.Println("nil-указатель:", inv.AddItem(weapon))

	// 1. Use
	fmt.Println("--- Use ---")
	fmt.Println(sword.Use())
	fmt.Println(shield.Use())
	fmt.Println(potion.Use())
	fmt.Println(bow.Use())

	// 2. DescribeItem
	fmt.Println("--- DescribeItem ---")
	fmt.Println(DescribeItem(sword))
	fmt.Println(DescribeItem(nil))
	fmt.Println(DescribeItem(weapon))

	// 3. Weapons
	fmt.Println("--- Weapons ---")
	for _, w := range inv.Weapons() {
		fmt.Printf("%s (урон: %d, прочность: %d)\n", w.Name(), w.Damage, w.Durability)
	}

	// 4. BrokenItems
	fmt.Println("--- BrokenItems ---")
	for _, item := range inv.BrokenItems() {
		fmt.Println(DescribeItem(item))
	}

	// 5. ItemNames
	fmt.Println("--- ItemNames ---")
	fmt.Println(strings.Join(inv.ItemNames(), ", "))

	// 6. FindItemByName
	fmt.Println("--- FindItemByName ---")
	for _, name := range []string{"Щит", "Копьё"} {
		if item, ok := inv.FindItemByName(name); ok {
			fmt.Println("Найдено:", DescribeItem(item))
		} else {
			fmt.Printf("Предмет %q не найден\n", name)
		}
	}

	// Бонус: сохранение и загрузка
	fmt.Println("--- Save / Load ---")

	// Имя с разделителем не должно попасть в файл
	var sb strings.Builder
	badSword := &Weapon{name: "Меч|2", weight: 1, Damage: 1, Durability: 1}
	fmt.Println("Имя с «|»:", badSword.Serialize(&sb))

	const path = "inventory.txt"
	if err := saveToFile(inv, path); err != nil {
		fmt.Println("Ошибка сохранения:", err)
		return
	}

	loaded := &Inventory{}
	if err := loadFromFile(loaded, path); err != nil {
		fmt.Println("Ошибка загрузки:", err)
		return
	}
	fmt.Println("Загружено:", strings.Join(loaded.ItemNames(), ", "))
}
