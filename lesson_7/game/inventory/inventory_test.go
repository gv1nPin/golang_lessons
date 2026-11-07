package inventory

import (
	"errors"
	"slices"
	"testing"
)

// entryIDs возвращает ID всех предметов в порядке хранения.
func entryIDs(inv *Inventory) []int {
	var ids []int
	for _, entry := range inv.Entries() {
		ids = append(ids, entry.ID)
	}
	return ids
}

// newThreeItems создаёт инвентарь с тремя предметами: ID 1, 2 и 3.
func newThreeItems(t *testing.T) *Inventory {
	t.Helper()
	inv := New()
	mustAdd(t, inv, mustWeapon(t, "Меч", 5))
	mustAdd(t, inv, mustArmor(t, "Щит"))
	mustAdd(t, inv, mustPotion(t, "Лечебное", 3))
	return inv
}

func TestAddItem(t *testing.T) {
	inv := New()

	for wantID, item := range []Item{mustWeapon(t, "Меч", 1), mustArmor(t, "Щит"), mustPotion(t, "Зелье", 1)} {
		id, err := inv.AddItem(item)
		if err != nil {
			t.Fatal(err)
		}
		if id != wantID+1 {
			t.Errorf("ID = %d, хочу %d", id, wantID+1)
		}
	}
	if got, want := entryIDs(inv), []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Errorf("ID в инвентаре = %v, хочу %v", got, want)
	}
}

func TestAddItemRejectsNil(t *testing.T) {
	var (
		nilWeapon *Weapon
		nilArmor  *Armor
		nilPotion *Potion
	)
	tests := []struct {
		name string
		item Item
	}{
		{"nil", nil},
		{"nil-указатель на оружие", nilWeapon},
		{"nil-указатель на броню", nilArmor},
		{"nil-указатель на зелье", nilPotion},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := New()

			id, err := inv.AddItem(tt.item)

			if !errors.Is(err, ErrNilItem) {
				t.Errorf("err = %v, хочу ErrNilItem", err)
			}
			if id != 0 {
				t.Errorf("ID = %d, хочу 0 при ошибке", id)
			}
			if n := len(inv.Entries()); n != 0 {
				t.Errorf("в инвентаре %d предметов, хочу 0", n)
			}
			// Отказ не должен тратить ID.
			if next := mustAdd(t, inv, mustWeapon(t, "Меч", 1)); next != 1 {
				t.Errorf("следующий ID = %d, хочу 1", next)
			}
		})
	}
}

func TestAddItemWithID(t *testing.T) {
	var nilWeapon *Weapon
	tests := []struct {
		name       string
		id         int
		item       Item
		wantErr    error
		wantNextID int // ID, который выдаст AddItem после этого вызова
	}{
		{"свободный ID сразу за последним", 3, mustWeapon(t, "Нож", 1), nil, 4},
		{"большой ID", 10, mustWeapon(t, "Нож", 1), nil, 11},
		{"занятый ID", 2, mustWeapon(t, "Нож", 1), ErrDuplicateID, 3},
		{"нулевой ID", 0, mustWeapon(t, "Нож", 1), ErrInvalidID, 3},
		{"отрицательный ID", -5, mustWeapon(t, "Нож", 1), ErrInvalidID, 3},
		{"nil вместо предмета", 7, nil, ErrNilItem, 3},
		{"nil-указатель вместо предмета", 7, nilWeapon, ErrNilItem, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := New()
			mustAdd(t, inv, mustWeapon(t, "Меч", 1)) // ID 1
			mustAdd(t, inv, mustArmor(t, "Щит"))     // ID 2

			err := inv.AddItemWithID(tt.id, tt.item)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, хочу %v", err, tt.wantErr)
			}
			wantEntries := 2
			if tt.wantErr == nil {
				wantEntries = 3
			}
			if got := len(inv.Entries()); got != wantEntries {
				t.Errorf("предметов = %d, хочу %d", got, wantEntries)
			}
			if tt.wantErr == nil {
				if _, getErr := inv.Get(tt.id); getErr != nil {
					t.Errorf("добавленный предмет не найден по ID %d: %v", tt.id, getErr)
				}
			}
			if next := mustAdd(t, inv, mustPotion(t, "Зелье", 1)); next != tt.wantNextID {
				t.Errorf("следующий ID = %d, хочу %d", next, tt.wantNextID)
			}
		})
	}
}

// Инвентарь, созданный без New, тоже работает: ID начинаются с 1.
func TestZeroValueInventory(t *testing.T) {
	t.Run("AddItem", func(t *testing.T) {
		var inv Inventory
		if id := mustAdd(t, &inv, mustWeapon(t, "Меч", 1)); id != 1 {
			t.Errorf("первый ID = %d, хочу 1", id)
		}
		if id := mustAdd(t, &inv, mustWeapon(t, "Лук", 1)); id != 2 {
			t.Errorf("второй ID = %d, хочу 2", id)
		}
	})
	t.Run("AddItemWithID", func(t *testing.T) {
		var inv Inventory
		if err := inv.AddItemWithID(7, mustWeapon(t, "Меч", 1)); err != nil {
			t.Fatal(err)
		}
		if id := mustAdd(t, &inv, mustWeapon(t, "Лук", 1)); id != 8 {
			t.Errorf("ID после AddItemWithID(7) = %d, хочу 8", id)
		}
	})
}

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		wantName string
		wantErr  error
	}{
		{"первый", 1, "Меч", nil},
		{"средний", 2, "Щит", nil},
		{"последний", 3, "Лечебное", nil},
		{"несуществующий", 99, "", ErrNotFound},
		{"нулевой", 0, "", ErrNotFound},
		{"отрицательный", -1, "", ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := newThreeItems(t)

			item, err := inv.Get(tt.id)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, хочу %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if item != nil {
					t.Errorf("при ошибке предмет должен быть nil, получено %v", item)
				}
				return
			}
			if item.Name() != tt.wantName {
				t.Errorf("название = %q, хочу %q", item.Name(), tt.wantName)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		wantErr error
		wantIDs []int
	}{
		{"первый", 1, nil, []int{2, 3}},
		{"средний", 2, nil, []int{1, 3}},
		{"последний", 3, nil, []int{1, 2}},
		{"несуществующий", 99, ErrNotFound, []int{1, 2, 3}},
		{"нулевой", 0, ErrNotFound, []int{1, 2, 3}},
		{"отрицательный", -1, ErrNotFound, []int{1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := newThreeItems(t)

			err := inv.Delete(tt.id)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, хочу %v", err, tt.wantErr)
			}
			if got := entryIDs(inv); !slices.Equal(got, tt.wantIDs) {
				t.Errorf("ID после удаления = %v, хочу %v", got, tt.wantIDs)
			}
			if tt.wantErr == nil {
				if _, getErr := inv.Get(tt.id); !errors.Is(getErr, ErrNotFound) {
					t.Errorf("Get удалённого: err = %v, хочу ErrNotFound", getErr)
				}
			}
		})
	}
}

// Свободный ID меньше следующего (например, освобождённый удалением) принимается,
// а счётчик следующего ID при этом не уменьшается.
func TestAddItemWithIDBelowNextID(t *testing.T) {
	inv := newThreeItems(t)
	if err := inv.Delete(2); err != nil {
		t.Fatal(err)
	}

	if err := inv.AddItemWithID(2, mustWeapon(t, "Нож", 1)); err != nil {
		t.Fatalf("свободный ID 2: err = %v, хочу nil", err)
	}

	if got, want := entryIDs(inv), []int{1, 3, 2}; !slices.Equal(got, want) {
		t.Errorf("ID в инвентаре = %v, хочу %v (порядок добавления)", got, want)
	}
	if next := mustAdd(t, inv, mustWeapon(t, "Лук", 1)); next != 4 {
		t.Errorf("следующий ID = %d, хочу 4", next)
	}
}

// ID не переиспользуются: после удаления последнего предмета новый получает следующий номер.
func TestIDsAreNotReusedAfterDelete(t *testing.T) {
	inv := newThreeItems(t)
	if err := inv.Delete(3); err != nil {
		t.Fatal(err)
	}

	if id := mustAdd(t, inv, mustWeapon(t, "Лук", 1)); id != 4 {
		t.Errorf("ID = %d, хочу 4", id)
	}
}

func TestEntriesReturnsCopy(t *testing.T) {
	inv := newThreeItems(t)

	entries := inv.Entries()
	entries[0] = Entry{}
	_ = append(entries, Entry{ID: 100})

	if got, want := entryIDs(inv), []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Errorf("изменение копии затронуло инвентарь: %v, хочу %v", got, want)
	}
}

// Use через Get меняет тот же предмет, который лежит в инвентаре.
func TestUseThroughGetChangesStoredItem(t *testing.T) {
	inv := New()
	id := mustAdd(t, inv, mustWeapon(t, "Меч", 2))

	item, err := inv.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := item.Use(); err != nil {
		t.Fatal(err)
	}

	again, err := inv.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if weapon := again.(*Weapon); weapon.Durability != 1 {
		t.Errorf("прочность = %d, хочу 1", weapon.Durability)
	}
}

func names(items []Item) []string {
	var result []string
	for _, item := range items {
		result = append(result, item.Name())
	}
	return result
}

func weaponNames(weapons []*Weapon) []string {
	var result []string
	for _, w := range weapons {
		result = append(result, w.Name())
	}
	return result
}

func TestQueries(t *testing.T) {
	tests := []struct {
		name        string
		items       []Item
		wantWeapons []string
		wantBroken  []string
		wantNames   []string
	}{
		{
			name: "пустой инвентарь",
		},
		{
			name:      "только броня",
			items:     []Item{mustArmor(t, "Щит")},
			wantNames: []string{"Щит"},
		},
		{
			name: "целое оружие и полное зелье",
			items: []Item{
				mustWeapon(t, "Меч", 5),
				mustPotion(t, "Лечебное", 3),
			},
			wantWeapons: []string{"Меч"},
			wantNames:   []string{"Меч", "Лечебное"},
		},
		{
			name: "граница: прочность и заряды ровно 0 или меньше",
			items: []Item{
				mustWeapon(t, "Меч", 5),
				mustWeapon(t, "Сломанный лук", 0),
				mustArmor(t, "Щит"),
				mustPotion(t, "Пустое зелье", 0),
				mustPotion(t, "Полное зелье", 3),
				mustWeapon(t, "Ржавый нож", -1),
				mustPotion(t, "Отрицательное зелье", -1),
			},
			wantWeapons: []string{"Меч", "Сломанный лук", "Ржавый нож"},
			wantBroken:  []string{"Сломанный лук", "Пустое зелье", "Ржавый нож", "Отрицательное зелье"},
			wantNames: []string{
				"Меч", "Сломанный лук", "Щит", "Пустое зелье", "Полное зелье", "Ржавый нож", "Отрицательное зелье",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := New()
			for _, item := range tt.items {
				mustAdd(t, inv, item)
			}

			if got := weaponNames(inv.Weapons()); !slices.Equal(got, tt.wantWeapons) {
				t.Errorf("Weapons = %v, хочу %v", got, tt.wantWeapons)
			}
			if got := names(inv.BrokenItems()); !slices.Equal(got, tt.wantBroken) {
				t.Errorf("BrokenItems = %v, хочу %v", got, tt.wantBroken)
			}
			if got := inv.ItemNames(); !slices.Equal(got, tt.wantNames) {
				t.Errorf("ItemNames = %v, хочу %v", got, tt.wantNames)
			}
		})
	}
}

// Меняя предмет, который вернул Use, запросы видят новое состояние.
func TestQueriesSeeStateChanges(t *testing.T) {
	inv := New()
	id := mustAdd(t, inv, mustWeapon(t, "Меч", 1))
	if len(inv.BrokenItems()) != 0 {
		t.Fatal("целый меч не должен считаться сломанным")
	}

	item, err := inv.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := item.Use(); err != nil {
		t.Fatal(err)
	}

	if got, want := names(inv.BrokenItems()), []string{"Меч"}; !slices.Equal(got, want) {
		t.Errorf("BrokenItems после износа = %v, хочу %v", got, want)
	}
}

func TestFindItemByName(t *testing.T) {
	first := mustWeapon(t, "Меч", 1)
	second := mustWeapon(t, "Меч", 2)
	shield := mustArmor(t, "Щит")

	inv := New()
	mustAdd(t, inv, first)
	mustAdd(t, inv, shield)
	mustAdd(t, inv, second)

	tests := []struct {
		name      string
		query     string
		want      Item
		wantFound bool
	}{
		{"существующий предмет", "Щит", shield, true},
		{"при дубликатах возвращается первый", "Меч", first, true},
		{"регистр важен", "меч", nil, false},
		{"пустое название", "", nil, false},
		{"несуществующий", "Копьё", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := inv.FindItemByName(tt.query)

			if found != tt.wantFound {
				t.Fatalf("found = %v, хочу %v", found, tt.wantFound)
			}
			if got != tt.want {
				t.Errorf("предмет = %v, хочу %v", got, tt.want)
			}
		})
	}
}

func TestFindItemByNameInEmptyInventory(t *testing.T) {
	if item, found := New().FindItemByName("Меч"); found || item != nil {
		t.Errorf("FindItemByName в пустом инвентаре = %v, %v", item, found)
	}
}
