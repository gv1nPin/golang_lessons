// Package inventory описывает предметы (оружие, броню, зелья) и инвентарь для их хранения.
// Пакет ничего не знает ни о файлах, ни о командной строке.
package inventory

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// Ошибки предметов. Вызывающий код различает их через errors.Is
// и никогда не сравнивает текст ошибки.
var (
	// ErrBroken возвращает Use у сломанного предмета.
	ErrBroken = errors.New("item is broken")
	// ErrExhausted возвращает Use у предмета, у которого закончились заряды.
	ErrExhausted = errors.New("item is exhausted")
	// ErrInvalidItem оборачивают конструкторы, когда им передали неверные значения.
	ErrInvalidItem = errors.New("invalid item")
)

// reservedChars — символы, которые нельзя использовать в названии и эффекте:
// они служат разделителями в файле хранения (поля через "|", одна запись на строку).
const reservedChars = "|\r\n"

// Item — любой предмет инвентаря.
type Item interface {
	Name() string
	Weight() float64
	// Use использует предмет. Успешное использование меняет состояние предмета.
	// Если предмет использовать нельзя, возвращается ошибка (ErrBroken, ErrExhausted).
	Use() (string, error)
}

// Weapon — оружие.
//
// name и weight с маленькой буквы: поле Name конфликтовало бы с методом Name().
// Из других пакетов оружие создаётся через NewWeapon.
type Weapon struct {
	name       string
	weight     float64
	Damage     int
	Durability int
}

// Armor — броня. Equipped показывает, надета ли она.
type Armor struct {
	name     string
	weight   float64
	Defense  int
	Equipped bool
}

// Potion — зелье.
type Potion struct {
	name    string
	weight  float64
	Effect  string
	Charges int
}

// Compile-time проверки: если метод забыт или имеет не ту сигнатуру,
// код не скомпилируется.
var (
	_ Item = (*Weapon)(nil)
	_ Item = (*Armor)(nil)
	_ Item = (*Potion)(nil)
)

// validate — единственное место, где проверяются название и вес предмета.
// Все предметы, и созданные в коде, и прочитанные из файла, проходят через конструкторы,
// поэтому в инвентаре не может оказаться предмета, который нельзя сохранить.
func validate(name string, weight float64) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: name must not be empty", ErrInvalidItem)
	}
	if err := checkText("name", name); err != nil {
		return err
	}
	if math.IsNaN(weight) || math.IsInf(weight, 0) || weight < 0 {
		return fmt.Errorf("%w: weight must be a finite non-negative number, got %v", ErrInvalidItem, weight)
	}
	return nil
}

// checkText отклоняет текст с зарезервированными символами.
func checkText(label, value string) error {
	if strings.ContainsAny(value, reservedChars) {
		return fmt.Errorf("%w: %s %q must not contain '|' or newlines", ErrInvalidItem, label, value)
	}
	return nil
}

// NewWeapon создаёт оружие. Название не может быть пустым или содержать '|' и переводы строк,
// вес должен быть конечным неотрицательным числом.
func NewWeapon(name string, weight float64, damage, durability int) (*Weapon, error) {
	if err := validate(name, weight); err != nil {
		return nil, err
	}
	return &Weapon{name: name, weight: weight, Damage: damage, Durability: durability}, nil
}

// NewArmor создаёт броню. Правила для названия и веса те же, что у NewWeapon.
func NewArmor(name string, weight float64, defense int) (*Armor, error) {
	if err := validate(name, weight); err != nil {
		return nil, err
	}
	return &Armor{name: name, weight: weight, Defense: defense}, nil
}

// NewPotion создаёт зелье. Правила для названия и веса те же, что у NewWeapon;
// эффект тоже не может содержать '|' и переводы строк.
func NewPotion(name string, weight float64, effect string, charges int) (*Potion, error) {
	if err := validate(name, weight); err != nil {
		return nil, err
	}
	if err := checkText("effect", effect); err != nil {
		return nil, err
	}
	return &Potion{name: name, weight: weight, Effect: effect, Charges: charges}, nil
}

// Name возвращает название оружия.
func (w *Weapon) Name() string { return w.name }

// Weight возвращает вес оружия.
func (w *Weapon) Weight() float64 { return w.weight }

// Use использует оружие и уменьшает его прочность на единицу.
// У сломанного оружия (прочность <= 0) возвращает ErrBroken и ничего не меняет.
func (w *Weapon) Use() (string, error) {
	if w.Durability <= 0 {
		return "", fmt.Errorf("%s: %w", w.name, ErrBroken)
	}
	w.Durability--
	return fmt.Sprintf("Используется %s (урон %d, прочность осталась %d)", w.name, w.Damage, w.Durability), nil
}

// Name возвращает название брони.
func (a *Armor) Name() string { return a.name }

// Weight возвращает вес брони.
func (a *Armor) Weight() float64 { return a.weight }

// Use надевает броню: первое использование меняет Equipped на true.
// Если броня уже надета, состояние не меняется, и это не ошибка.
func (a *Armor) Use() (string, error) {
	if a.Equipped {
		return fmt.Sprintf("%s уже надето", a.name), nil
	}
	a.Equipped = true
	return fmt.Sprintf("Надевается %s (защита %d)", a.name, a.Defense), nil
}

// Name возвращает название зелья.
func (p *Potion) Name() string { return p.name }

// Weight возвращает вес зелья.
func (p *Potion) Weight() float64 { return p.weight }

// Use использует зелье и уменьшает число зарядов на единицу.
// У пустого зелья (заряды <= 0) возвращает ErrExhausted и ничего не меняет.
func (p *Potion) Use() (string, error) {
	if p.Charges <= 0 {
		return "", fmt.Errorf("%s: %w", p.name, ErrExhausted)
	}
	p.Charges--
	return fmt.Sprintf("Выпито %s (%s), зарядов осталось %d", p.name, p.Effect, p.Charges), nil
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

// DescribeItem возвращает краткое описание предмета: "<название> (вес: <вес>)".
// Для nil и для nil-указателя возвращает "Предмет отсутствует".
func DescribeItem(i Item) string {
	if isNilItem(i) {
		return "Предмет отсутствует"
	}
	return fmt.Sprintf("%s (вес: %g)", i.Name(), i.Weight())
}
