package inventory

import (
	"errors"
	"math"
	"strings"
	"testing"
)

// Помощники для всех тестовых файлов пакета.

func mustWeapon(t *testing.T, name string, durability int) *Weapon {
	t.Helper()
	w, err := NewWeapon(name, 1, 1, durability)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func mustArmor(t *testing.T, name string) *Armor {
	t.Helper()
	a, err := NewArmor(name, 4.5, 5)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func mustPotion(t *testing.T, name string, charges int) *Potion {
	t.Helper()
	p, err := NewPotion(name, 0.3, "+50 HP", charges)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func mustAdd(t *testing.T, inv *Inventory, item Item) int {
	t.Helper()
	id, err := inv.AddItem(item)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// Тестовые предметы: проверяют поведение кода на предметах, которых нет в пакете.

// panickyItem — предмет, у которого Use паникует.
type panickyItem struct{}

func (panickyItem) Name() string         { return "Ловушка" }
func (panickyItem) Weight() float64      { return 1 }
func (panickyItem) Use() (string, error) { panic("ловушка сработала") }

// errorItem — предмет, у которого Use всегда возвращает заданную ошибку.
type errorItem struct{ err error }

func (errorItem) Name() string           { return "Капризный" }
func (errorItem) Weight() float64        { return 1 }
func (e errorItem) Use() (string, error) { return "", e.err }

func TestNewWeapon(t *testing.T) {
	valid := []struct {
		name       string
		weight     float64
		damage     int
		durability int
	}{
		{"Меч", 3.5, 10, 5},
		{"Нулевой вес", 0, 1, 1},
		{"Нулевые характеристики", 1, 0, 0},
		{"Отрицательные характеристики", 1, -1, -1},
	}
	for _, tt := range valid {
		t.Run("корректно/"+tt.name, func(t *testing.T) {
			w, err := NewWeapon(tt.name, tt.weight, tt.damage, tt.durability)
			if err != nil {
				t.Fatal(err)
			}
			if w.Name() != tt.name || w.Weight() != tt.weight || w.Damage != tt.damage || w.Durability != tt.durability {
				t.Errorf("создано %+v, хочу %q, вес %v, урон %d, прочность %d",
					*w, tt.name, tt.weight, tt.damage, tt.durability)
			}
		})
	}
}

func TestNewArmor(t *testing.T) {
	armor, err := NewArmor("Щит", 4.5, 5)
	if err != nil {
		t.Fatal(err)
	}
	if armor.Name() != "Щит" || armor.Weight() != 4.5 || armor.Defense != 5 {
		t.Errorf("создано %+v", *armor)
	}
	if armor.Equipped {
		t.Error("новая броня не должна быть надета")
	}
}

func TestNewPotion(t *testing.T) {
	potion, err := NewPotion("Лечебное", 0.3, "+50 HP", 3)
	if err != nil {
		t.Fatal(err)
	}
	if potion.Name() != "Лечебное" || potion.Weight() != 0.3 || potion.Effect != "+50 HP" || potion.Charges != 3 {
		t.Errorf("создано %+v", *potion)
	}

	// Пустой эффект допустим: правила касаются только зарезервированных символов.
	if _, err := NewPotion("Вода", 0.1, "", 1); err != nil {
		t.Errorf("пустой эффект: err = %v, хочу nil", err)
	}
}

// Каждый конструктор отклоняет неверные значения ошибкой ErrInvalidItem и возвращает nil.
func TestConstructorsRejectInvalidValues(t *testing.T) {
	weapon := func(name string, weight float64) func() (bool, error) {
		return func() (bool, error) {
			w, err := NewWeapon(name, weight, 1, 1)
			return w == nil, err
		}
	}
	armor := func(name string, weight float64) func() (bool, error) {
		return func() (bool, error) {
			a, err := NewArmor(name, weight, 1)
			return a == nil, err
		}
	}
	potion := func(name string, weight float64, effect string) func() (bool, error) {
		return func() (bool, error) {
			p, err := NewPotion(name, weight, effect, 1)
			return p == nil, err
		}
	}

	tests := []struct {
		name  string
		build func() (isNil bool, err error)
	}{
		{"оружие: пустое название", weapon("", 1)},
		{"оружие: название из пробелов", weapon("   ", 1)},
		{"оружие: разделитель в названии", weapon("Меч|2", 1)},
		{"оружие: перевод строки в названии", weapon("Меч\n", 1)},
		{"оружие: возврат каретки в названии", weapon("Меч\r", 1)},
		{"оружие: отрицательный вес", weapon("Меч", -0.1)},
		{"оружие: NaN", weapon("Меч", math.NaN())},
		{"оружие: +Inf", weapon("Меч", math.Inf(1))},
		{"оружие: -Inf", weapon("Меч", math.Inf(-1))},
		{"броня: пустое название", armor("", 1)},
		{"броня: разделитель в названии", armor("Щит|2", 1)},
		{"броня: отрицательный вес", armor("Щит", -1)},
		{"зелье: пустое название", potion("", 1, "x")},
		{"зелье: разделитель в названии", potion("Зелье|2", 1, "x")},
		{"зелье: отрицательный вес", potion("Зелье", -1, "x")},
		{"зелье: разделитель в эффекте", potion("Зелье", 1, "a|b")},
		{"зелье: перевод строки в эффекте", potion("Зелье", 1, "a\nb")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isNil, err := tt.build()
			if !errors.Is(err, ErrInvalidItem) {
				t.Errorf("err = %v, хочу ErrInvalidItem", err)
			}
			if !isNil {
				t.Error("вместе с ошибкой вернулся не nil-предмет")
			}
		})
	}
}

func TestWeaponUse(t *testing.T) {
	tests := []struct {
		name           string
		durability     int
		wantErr        error
		wantDurability int
		wantMessage    string // подстрока сообщения при успехе
	}{
		{"прочности достаточно", 2, nil, 1, "прочность осталась 1"},
		{"последнее использование", 1, nil, 0, "прочность осталась 0"},
		{"сломано: прочность 0", 0, ErrBroken, 0, ""},
		{"сломано: отрицательная прочность", -3, ErrBroken, -3, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := mustWeapon(t, "Меч", tt.durability)

			message, err := w.Use()

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, хочу %v", err, tt.wantErr)
			}
			if w.Durability != tt.wantDurability {
				t.Errorf("прочность = %d, хочу %d", w.Durability, tt.wantDurability)
			}
			if tt.wantErr != nil {
				if message != "" {
					t.Errorf("при ошибке сообщение должно быть пустым, получено %q", message)
				}
				if !strings.Contains(err.Error(), "Меч") {
					t.Errorf("в тексте ошибки нет названия предмета: %q", err)
				}
				return
			}
			if !strings.Contains(message, "Меч") || !strings.Contains(message, tt.wantMessage) {
				t.Errorf("сообщение = %q, хочу название и %q", message, tt.wantMessage)
			}
		})
	}
}

// Оружие изнашивается шаг за шагом и ломается ровно после последнего использования.
func TestWeaponWearsOutStepByStep(t *testing.T) {
	w := mustWeapon(t, "Меч", 3)

	for wantDurability := 2; wantDurability >= 0; wantDurability-- {
		if _, err := w.Use(); err != nil {
			t.Fatalf("прочность была %d: %v", wantDurability+1, err)
		}
		if w.Durability != wantDurability {
			t.Fatalf("прочность = %d, хочу %d", w.Durability, wantDurability)
		}
	}
	if _, err := w.Use(); !errors.Is(err, ErrBroken) {
		t.Errorf("после износа err = %v, хочу ErrBroken", err)
	}
}

func TestPotionUse(t *testing.T) {
	tests := []struct {
		name        string
		charges     int
		wantErr     error
		wantCharges int
		wantMessage string
	}{
		{"заряды есть", 3, nil, 2, "зарядов осталось 2"},
		{"последний заряд", 1, nil, 0, "зарядов осталось 0"},
		{"пустое: заряды 0", 0, ErrExhausted, 0, ""},
		{"пустое: отрицательные заряды", -1, ErrExhausted, -1, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := mustPotion(t, "Лечебное", tt.charges)

			message, err := p.Use()

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, хочу %v", err, tt.wantErr)
			}
			if p.Charges != tt.wantCharges {
				t.Errorf("заряды = %d, хочу %d", p.Charges, tt.wantCharges)
			}
			if tt.wantErr != nil {
				if message != "" {
					t.Errorf("при ошибке сообщение должно быть пустым, получено %q", message)
				}
				return
			}
			if !strings.Contains(message, "+50 HP") || !strings.Contains(message, tt.wantMessage) {
				t.Errorf("сообщение = %q, хочу эффект и %q", message, tt.wantMessage)
			}
		})
	}
}

// Ошибки использования различаются только через errors.Is: ErrBroken и ErrExhausted не путаются.
func TestUseErrorsAreDistinct(t *testing.T) {
	_, brokenErr := mustWeapon(t, "Меч", 0).Use()
	_, exhaustedErr := mustPotion(t, "Зелье", 0).Use()

	if errors.Is(brokenErr, ErrExhausted) {
		t.Error("ошибка сломанного оружия не должна быть ErrExhausted")
	}
	if errors.Is(exhaustedErr, ErrBroken) {
		t.Error("ошибка пустого зелья не должна быть ErrBroken")
	}
}

func TestArmorUse(t *testing.T) {
	a := mustArmor(t, "Щит")

	steps := []struct {
		name        string
		wantMessage string
	}{
		{"первое использование надевает", "Надевается Щит"},
		{"повторное использование ничего не меняет", "Щит уже надето"},
		{"и ещё раз", "Щит уже надето"},
	}
	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			message, err := a.Use()
			if err != nil {
				t.Fatalf("err = %v, хочу nil", err)
			}
			if !strings.Contains(message, step.wantMessage) {
				t.Errorf("сообщение = %q, хочу %q", message, step.wantMessage)
			}
			if !a.Equipped {
				t.Error("после Use броня должна быть надета")
			}
		})
	}
}

func TestDescribeItem(t *testing.T) {
	var (
		nilWeapon *Weapon
		nilArmor  *Armor
		nilPotion *Potion
	)
	tests := []struct {
		name string
		item Item
		want string
	}{
		{"оружие", mustWeapon(t, "Меч", 1), "Меч (вес: 1)"},
		{"броня", mustArmor(t, "Щит"), "Щит (вес: 4.5)"},
		{"зелье", mustPotion(t, "Лечебное", 1), "Лечебное (вес: 0.3)"},
		{"nil", nil, "Предмет отсутствует"},
		{"nil-указатель на оружие", nilWeapon, "Предмет отсутствует"},
		{"nil-указатель на броню", nilArmor, "Предмет отсутствует"},
		{"nil-указатель на зелье", nilPotion, "Предмет отсутствует"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DescribeItem(tt.item); got != tt.want {
				t.Errorf("DescribeItem = %q, хочу %q", got, tt.want)
			}
		})
	}
}
