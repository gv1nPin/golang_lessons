package inventory

import (
	"errors"
	"math"
	"slices"
	"strings"
	"testing"
)

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

func TestWeaponUse(t *testing.T) {
	w := mustWeapon(t, "Меч", 1)

	if _, err := w.Use(); err != nil {
		t.Fatalf("первое использование: %v", err)
	}
	if w.Durability != 0 {
		t.Errorf("прочность после использования = %d, хочу 0", w.Durability)
	}

	if _, err := w.Use(); !errors.Is(err, ErrBroken) {
		t.Errorf("err = %v, хочу ErrBroken", err)
	}
	if w.Durability != 0 {
		t.Errorf("неудачное использование изменило прочность: %d", w.Durability)
	}
}

func TestPotionUse(t *testing.T) {
	p := mustPotion(t, "Лечебное", 1)

	if _, err := p.Use(); err != nil {
		t.Fatalf("первое использование: %v", err)
	}
	if p.Charges != 0 {
		t.Errorf("заряды после использования = %d, хочу 0", p.Charges)
	}

	if _, err := p.Use(); !errors.Is(err, ErrExhausted) {
		t.Errorf("err = %v, хочу ErrExhausted", err)
	}
}

func TestArmorUseChangesState(t *testing.T) {
	a := mustArmor(t, "Щит")
	if a.Equipped {
		t.Fatal("новая броня не должна быть надета")
	}

	if _, err := a.Use(); err != nil {
		t.Fatalf("первое использование: %v", err)
	}
	if !a.Equipped {
		t.Error("после Use броня должна быть надета")
	}

	message, err := a.Use()
	if err != nil {
		t.Fatalf("повторное использование: %v", err)
	}
	if !strings.Contains(message, "уже надето") {
		t.Errorf("сообщение = %q, хочу упоминание «уже надето»", message)
	}
}

func TestConstructorsRejectInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		build func() error
	}{
		{"пустое название", func() error { _, err := NewWeapon("  ", 1, 1, 1); return err }},
		{"разделитель в названии", func() error { _, err := NewArmor("Щит|2", 1, 1); return err }},
		{"перевод строки в названии", func() error { _, err := NewPotion("Зелье\n", 1, "x", 1); return err }},
		{"разделитель в эффекте", func() error { _, err := NewPotion("Зелье", 1, "a|b", 1); return err }},
		{"отрицательный вес", func() error { _, err := NewWeapon("Меч", -1, 1, 1); return err }},
		{"NaN в весе", func() error { _, err := NewWeapon("Меч", math.NaN(), 1, 1); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.build(); !errors.Is(err, ErrInvalidItem) {
				t.Errorf("err = %v, хочу ErrInvalidItem", err)
			}
		})
	}
}

func TestInventoryIDs(t *testing.T) {
	inv := New()

	id1 := mustAdd(t, inv, mustWeapon(t, "Меч", 1))
	id2 := mustAdd(t, inv, mustWeapon(t, "Лук", 1))
	if id1 != 1 || id2 != 2 {
		t.Fatalf("ID = %d и %d, хочу 1 и 2", id1, id2)
	}

	if err := inv.Delete(id1); err != nil {
		t.Fatal(err)
	}
	if _, err := inv.Get(id1); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get удалённого: err = %v, хочу ErrNotFound", err)
	}
	if err := inv.Delete(id1); !errors.Is(err, ErrNotFound) {
		t.Errorf("повторный Delete: err = %v, хочу ErrNotFound", err)
	}

	if id3 := mustAdd(t, inv, mustWeapon(t, "Копьё", 1)); id3 != 3 {
		t.Errorf("ID нового предмета = %d, хочу 3 (ID не переиспользуются)", id3)
	}

	if err := inv.AddItemWithID(2, mustWeapon(t, "Нож", 1)); !errors.Is(err, ErrDuplicateID) {
		t.Errorf("дубликат ID: err = %v, хочу ErrDuplicateID", err)
	}
	if err := inv.AddItemWithID(0, mustWeapon(t, "Нож", 1)); !errors.Is(err, ErrInvalidID) {
		t.Errorf("ID 0: err = %v, хочу ErrInvalidID", err)
	}

	if err := inv.AddItemWithID(10, mustWeapon(t, "Нож", 1)); err != nil {
		t.Fatal(err)
	}
	if id := mustAdd(t, inv, mustWeapon(t, "Топор", 1)); id != 11 {
		t.Errorf("ID после AddItemWithID(10) = %d, хочу 11", id)
	}
}

func TestAddItemRejectsNil(t *testing.T) {
	inv := New()

	if _, err := inv.AddItem(nil); !errors.Is(err, ErrNilItem) {
		t.Errorf("nil: err = %v, хочу ErrNilItem", err)
	}
	var weapon *Weapon // nil-указатель внутри интерфейса
	if _, err := inv.AddItem(weapon); !errors.Is(err, ErrNilItem) {
		t.Errorf("nil-указатель: err = %v, хочу ErrNilItem", err)
	}
	if n := len(inv.Entries()); n != 0 {
		t.Errorf("в инвентаре %d предметов, хочу 0", n)
	}
}

func TestQueries(t *testing.T) {
	inv := New()
	mustAdd(t, inv, mustWeapon(t, "Меч", 5))
	mustAdd(t, inv, mustWeapon(t, "Сломанный лук", 0))
	mustAdd(t, inv, mustArmor(t, "Щит"))
	mustAdd(t, inv, mustPotion(t, "Пустое зелье", 0))

	if got := len(inv.Weapons()); got != 2 {
		t.Errorf("оружия = %d, хочу 2", got)
	}

	var broken []string
	for _, item := range inv.BrokenItems() {
		broken = append(broken, item.Name())
	}
	if want := []string{"Сломанный лук", "Пустое зелье"}; !slices.Equal(broken, want) {
		t.Errorf("сломанные = %v, хочу %v", broken, want)
	}

	wantNames := []string{"Меч", "Сломанный лук", "Щит", "Пустое зелье"}
	if got := inv.ItemNames(); !slices.Equal(got, wantNames) {
		t.Errorf("названия = %v, хочу %v", got, wantNames)
	}

	if item, ok := inv.FindItemByName("Щит"); !ok || item.Name() != "Щит" {
		t.Errorf("FindItemByName(Щит) = %v, %v", item, ok)
	}
	if _, ok := inv.FindItemByName("Копьё"); ok {
		t.Error("FindItemByName(Копьё) нашёл несуществующий предмет")
	}
}

func TestDescribeItem(t *testing.T) {
	sword, err := NewWeapon("Меч", 3.5, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := DescribeItem(sword), "Меч (вес: 3.5)"; got != want {
		t.Errorf("DescribeItem(меч) = %q, хочу %q", got, want)
	}
	if got := DescribeItem(nil); got != "Предмет отсутствует" {
		t.Errorf("DescribeItem(nil) = %q", got)
	}
	var nilWeapon *Weapon
	if got := DescribeItem(nilWeapon); got != "Предмет отсутствует" {
		t.Errorf("DescribeItem(nil-указатель) = %q", got)
	}
}

// panickyItem — предмет, у которого Use паникует: проверяет, что UseDangerousItem
// перехватывает не только собственную панику.
type panickyItem struct{}

func (panickyItem) Name() string          { return "Ловушка" }
func (panickyItem) Weight() float64       { return 1 }
func (panickyItem) Use() (string, error) { panic("ловушка сработала") }

func TestUseDangerousItem(t *testing.T) {
	bomb := mustWeapon(t, "Бомба", 1)
	message, err := UseDangerousItem(bomb)
	if !errors.Is(err, ErrPanic) {
		t.Errorf("бомба: err = %v, хочу ErrPanic", err)
	}
	if message != "" {
		t.Errorf("message = %q, хочу пустую строку", message)
	}
	if bomb.Durability != 1 {
		t.Errorf("после паники Use не должен был вызываться, прочность = %d", bomb.Durability)
	}

	if _, err := UseDangerousItem(panickyItem{}); !errors.Is(err, ErrPanic) {
		t.Errorf("паника внутри Use: err = %v, хочу ErrPanic", err)
	}

	// Обычные проблемы остаются обычными ошибками.
	if _, err := UseDangerousItem(nil); !errors.Is(err, ErrNilItem) {
		t.Errorf("nil: err = %v, хочу ErrNilItem", err)
	}
	var nilWeapon *Weapon
	if _, err := UseDangerousItem(nilWeapon); !errors.Is(err, ErrNilItem) {
		t.Errorf("nil-указатель: err = %v, хочу ErrNilItem", err)
	}

	sword := mustWeapon(t, "Меч", 1)
	if _, err := UseDangerousItem(sword); err != nil {
		t.Errorf("обычное использование: %v", err)
	}
	_, err = UseDangerousItem(sword)
	if !errors.Is(err, ErrBroken) || errors.Is(err, ErrPanic) {
		t.Errorf("сломанный предмет: err = %v, хочу ErrBroken, а не ErrPanic", err)
	}
}
