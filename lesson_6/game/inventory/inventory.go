package inventory

import (
	"errors"
	"fmt"
)

// Ошибки инвентаря. Проверяются через errors.Is.
var (
	// ErrNilItem возвращают методы добавления, когда им передали пустой предмет.
	ErrNilItem = errors.New("item is nil")
	// ErrNotFound возвращают Get и Delete, когда предмета с таким ID нет.
	ErrNotFound = errors.New("item not found")
	// ErrDuplicateID возвращает AddItemWithID, когда предмет с таким ID уже есть.
	ErrDuplicateID = errors.New("duplicate item id")
	// ErrInvalidID возвращает AddItemWithID для ID меньше 1.
	ErrInvalidID = errors.New("invalid item id")
)

// Entry — предмет вместе с его числовым ID. ID принадлежит инвентарю, а не предмету:
// один и тот же предмет не может получить два разных ID.
type Entry struct {
	ID   int
	Item Item
}

// Inventory — набор предметов. Каждому предмету присваивается числовой ID, начиная с 1.
type Inventory struct {
	entries []Entry
	nextID  int
}

// New создаёт пустой инвентарь.
func New() *Inventory {
	return &Inventory{nextID: 1}
}

// AddItem добавляет предмет и возвращает его ID. Пустые предметы (nil и nil-указатель)
// отклоняются: иначе методы вроде ItemNames упали бы с паникой.
func (inv *Inventory) AddItem(item Item) (int, error) {
	if isNilItem(item) {
		return 0, ErrNilItem
	}
	if inv.nextID < 1 {
		inv.nextID = 1
	}
	id := inv.nextID
	inv.entries = append(inv.entries, Entry{ID: id, Item: item})
	inv.nextID++
	return id, nil
}

// AddItemWithID добавляет предмет с заданным ID. Нужен при загрузке из хранилища,
// чтобы ID предметов не менялись между запусками. Следующий ID, который выдаст AddItem,
// будет больше любого из существующих.
func (inv *Inventory) AddItemWithID(id int, item Item) error {
	if isNilItem(item) {
		return ErrNilItem
	}
	if id < 1 {
		return fmt.Errorf("id %d: %w", id, ErrInvalidID)
	}
	if _, found := inv.indexOf(id); found {
		return fmt.Errorf("id %d: %w", id, ErrDuplicateID)
	}
	inv.entries = append(inv.entries, Entry{ID: id, Item: item})
	if id >= inv.nextID {
		inv.nextID = id + 1
	}
	return nil
}

// Entries возвращает копию списка предметов с их ID в порядке добавления.
func (inv *Inventory) Entries() []Entry {
	return append([]Entry(nil), inv.entries...)
}

// Get возвращает предмет по ID или ErrNotFound.
func (inv *Inventory) Get(id int) (Item, error) {
	i, found := inv.indexOf(id)
	if !found {
		return nil, fmt.Errorf("id %d: %w", id, ErrNotFound)
	}
	return inv.entries[i].Item, nil
}

// Delete удаляет предмет по ID или возвращает ErrNotFound.
func (inv *Inventory) Delete(id int) error {
	i, found := inv.indexOf(id)
	if !found {
		return fmt.Errorf("id %d: %w", id, ErrNotFound)
	}
	inv.entries = append(inv.entries[:i], inv.entries[i+1:]...)
	return nil
}

// indexOf возвращает позицию предмета с данным ID.
func (inv *Inventory) indexOf(id int) (int, bool) {
	for i, entry := range inv.entries {
		if entry.ID == id {
			return i, true
		}
	}
	return 0, false
}

// items возвращает предметы без ID.
func (inv *Inventory) items() []Item {
	return mapSlice(inv.entries, func(e Entry) Item { return e.Item })
}

// Weapons возвращает только оружие.
func (inv *Inventory) Weapons() []*Weapon {
	weaponItems := filter(inv.items(), func(item Item) bool {
		_, ok := item.(*Weapon)
		return ok
	})
	return mapSlice(weaponItems, func(item Item) *Weapon {
		return item.(*Weapon)
	})
}

// BrokenItems возвращает сломанное оружие (прочность <= 0) и пустые зелья (заряды <= 0).
func (inv *Inventory) BrokenItems() []Item {
	return filter(inv.items(), func(item Item) bool {
		switch v := item.(type) {
		case *Weapon:
			return v.Durability <= 0
		case *Potion:
			return v.Charges <= 0
		}
		return false
	})
}

// ItemNames возвращает названия всех предметов.
func (inv *Inventory) ItemNames() []string {
	return mapSlice(inv.entries, func(e Entry) string {
		return e.Item.Name()
	})
}

// FindItemByName ищет предмет по названию.
func (inv *Inventory) FindItemByName(name string) (Item, bool) {
	entry, ok := find(inv.entries, func(e Entry) bool {
		return e.Item.Name() == name
	})
	if !ok {
		return nil, false
	}
	return entry.Item, true
}
