package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang_lessons/lesson_7/game/inventory"
)

// store — то, что общее у обоих хранилищ.
type store interface {
	Load() (*inventory.Inventory, error)
	Save(inv *inventory.Inventory) error
}

var (
	_ store = (*FileStore)(nil)
	_ store = (*MemoryStore)(nil)
)

// storeFactory создаёт хранилище с готовым содержимым content (пустая строка — пустые данные).
type storeFactory func(t *testing.T, content string) store

func fileFactory(t *testing.T, content string) store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "inventory.dat")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return NewFileStore(path)
}

func memoryFactory(t *testing.T, content string) store {
	t.Helper()
	return NewMemoryStoreFrom(content)
}

func mustLoad(t *testing.T, s store) *inventory.Inventory {
	t.Helper()
	inv, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	return inv
}

func mustSave(t *testing.T, s store, inv *inventory.Inventory) {
	t.Helper()
	if err := s.Save(inv); err != nil {
		t.Fatal(err)
	}
}

func deleteIDs(ids ...int) func(*testing.T, *inventory.Inventory) {
	return func(t *testing.T, inv *inventory.Inventory) {
		t.Helper()
		for _, id := range ids {
			if err := inv.Delete(id); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func addWeapon(name string) func(*testing.T, *inventory.Inventory) {
	return func(t *testing.T, inv *inventory.Inventory) {
		t.Helper()
		if _, err := inv.AddItem(mustWeapon(t, name, 1.2, 5, 0)); err != nil {
			t.Fatal(err)
		}
	}
}

func nothing(*testing.T, *inventory.Inventory) {}

// TestStoreContract запускает одни и те же проверки для обоих хранилищ: правила сохранения,
// загрузки и обработки ошибок у них должны совпадать.
func TestStoreContract(t *testing.T) {
	factories := map[string]storeFactory{
		"file":   fileFactory,
		"memory": memoryFactory,
	}

	for name, factory := range factories {
		t.Run(name, func(t *testing.T) {
			t.Run("Load", func(t *testing.T) { contractLoad(t, factory) })
			t.Run("Save", func(t *testing.T) { contractSave(t, factory) })
			t.Run("ошибки Save не портят данные", func(t *testing.T) { contractSaveFailures(t, factory) })
			t.Run("состояние предметов", func(t *testing.T) { contractItemState(t, factory) })
			t.Run("независимость копий", func(t *testing.T) { contractIndependence(t, factory) })
		})
	}
}

func contractLoad(t *testing.T, factory storeFactory) {
	const (
		corrupted = "Weapon|1|Меч|10|5|3.5\nArmor|2|Щит|abc|4.5|0\nPotion|3|Лечебное|+50 HP|3|0.3\n"
		duplicate = "Weapon|1|Меч|10|5|3.5\nWeapon|1|Лук|5|0|1.2\n"
	)
	tests := []struct {
		name     string
		content  string
		want     string // содержимое после загрузки (для успешных случаев)
		wantLine int    // номер строки в ParseItemError; 0 — ошибки быть не должно
		wantIs   error
	}{
		{name: "пустые данные", content: "", want: ""},
		{name: "обычные данные", content: sampleDump, want: sampleDump},
		{name: "старая запись брони без Equipped", content: "Armor|1|Щит|5|4.5\n", want: "Armor|1|Щит|5|4.5|0\n"},
		{name: "повреждённая строка", content: corrupted, wantLine: 2},
		{name: "повторяющийся ID", content: duplicate, wantLine: 2, wantIs: inventory.ErrDuplicateID},
		{name: "нулевой ID", content: "Weapon|0|Меч|1|1|1\n", wantLine: 1, wantIs: inventory.ErrInvalidID},
		{name: "неверное значение", content: "Weapon|1|Меч|10|5|-1\n", wantLine: 1, wantIs: inventory.ErrInvalidItem},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv, err := factory(t, tt.content).Load()

			if tt.wantLine == 0 {
				if err != nil {
					t.Fatal(err)
				}
				if got := dump(t, inv); got != tt.want {
					t.Errorf("загружено %q, хочу %q", got, tt.want)
				}
				return
			}

			if inv != nil {
				t.Error("при ошибке инвентарь не должен возвращаться")
			}
			var parseErr *ParseItemError
			if !errors.As(err, &parseErr) {
				t.Fatalf("err = %v, хочу *ParseItemError", err)
			}
			if parseErr.Line != tt.wantLine {
				t.Errorf("Line = %d, хочу %d", parseErr.Line, tt.wantLine)
			}
			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Errorf("err = %v, хочу %v", err, tt.wantIs)
			}
		})
	}
}

// contractSave последовательно сохраняет изменяющийся инвентарь в одно хранилище
// и после каждого шага проверяет, что загрузка возвращает ровно то, что сохранено.
func contractSave(t *testing.T, factory storeFactory) {
	s := factory(t, "")
	inv := sampleInventory(t)

	steps := []struct {
		name   string
		change func(*testing.T, *inventory.Inventory)
		want   string
	}{
		{"первое сохранение", nothing, sampleDump},
		{"повторное сохранение того же инвентаря", nothing, sampleDump},
		{"новое содержимое короче старого", deleteIDs(2, 3), "Weapon|1|Меч|10|5|3.5\n"},
		{"пустой инвентарь", deleteIDs(1), ""},
		{"предмет после опустошения получает новый ID", addWeapon("Лук"), "Weapon|4|Лук|5|0|1.2\n"},
	}
	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			step.change(t, inv)
			mustSave(t, s, inv)

			if got := dump(t, mustLoad(t, s)); got != step.want {
				t.Errorf("загружено %q, хочу %q", got, step.want)
			}
		})
	}

	t.Run("следующий ID после загрузки продолжает нумерацию", func(t *testing.T) {
		loaded := mustLoad(t, factory(t, sampleDump))
		id, err := loaded.AddItem(mustWeapon(t, "Копьё", 2, 1, 1))
		if err != nil || id != 4 {
			t.Errorf("следующий ID = %d (err %v), хочу 4", id, err)
		}
	})
}

// contractSaveFailures: если Save не удался, прежние данные остаются как были.
func contractSaveFailures(t *testing.T, factory storeFactory) {
	tests := []struct {
		name string
		bad  func(t *testing.T) *inventory.Inventory
	}{
		{"единственный предмет неизвестного типа", func(t *testing.T) *inventory.Inventory {
			return newInventory(t, fakeItem{})
		}},
		{"неизвестный предмет после обычного", func(t *testing.T) *inventory.Inventory {
			return newInventory(t, mustWeapon(t, "Меч", 1, 1, 1), fakeItem{})
		}},
		{"неизвестный предмет перед обычным", func(t *testing.T) *inventory.Inventory {
			return newInventory(t, fakeItem{}, mustWeapon(t, "Меч", 1, 1, 1))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := factory(t, "")
			mustSave(t, s, sampleInventory(t))

			err := s.Save(tt.bad(t))

			if !errors.Is(err, ErrNotStorable) {
				t.Fatalf("err = %v, хочу ErrNotStorable", err)
			}
			if got := dump(t, mustLoad(t, s)); got != sampleDump {
				t.Errorf("после неудачного Save данные изменились: %q", got)
			}
		})
	}
}

// contractItemState: изменённое состояние предметов (после Use) переживает Save и Load.
func contractItemState(t *testing.T, factory storeFactory) {
	weapon := mustWeapon(t, "Меч", 3.5, 10, 2)
	armor := mustArmor(t, "Щит", 4.5, 5, false)
	potion := mustPotion(t, "Лечебное", 0.3, "+50 HP", 1)
	inv := newInventory(t, weapon, armor, potion)
	for _, item := range []inventory.Item{weapon, armor, potion} {
		if _, err := item.Use(); err != nil {
			t.Fatal(err)
		}
	}

	s := factory(t, "")
	mustSave(t, s, inv)
	loaded := mustLoad(t, s)

	checks := []struct {
		name  string
		id    int
		check func(item inventory.Item) bool
	}{
		{"прочность оружия уменьшилась", 1, func(item inventory.Item) bool {
			w, ok := item.(*inventory.Weapon)
			return ok && w.Durability == 1 && w.Damage == 10
		}},
		{"броня надета", 2, func(item inventory.Item) bool {
			a, ok := item.(*inventory.Armor)
			return ok && a.Equipped && a.Defense == 5
		}},
		{"заряды зелья кончились", 3, func(item inventory.Item) bool {
			p, ok := item.(*inventory.Potion)
			return ok && p.Charges == 0 && p.Effect == "+50 HP"
		}},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			item, err := loaded.Get(c.id)
			if err != nil {
				t.Fatal(err)
			}
			if !c.check(item) {
				t.Errorf("предмет после загрузки: %#v", item)
			}
		})
	}
}

// contractIndependence: хранилище не делит состояние с вызывающим кодом.
func contractIndependence(t *testing.T, factory storeFactory) {
	t.Run("изменения инвентаря после Save не попадают в хранилище", func(t *testing.T) {
		s := factory(t, "")
		inv := sampleInventory(t)
		mustSave(t, s, inv)

		deleteIDs(1)(t, inv)
		armor, err := inv.Get(2)
		if err != nil {
			t.Fatal(err)
		}
		armor.(*inventory.Armor).Equipped = false

		if got := dump(t, mustLoad(t, s)); got != sampleDump {
			t.Errorf("хранилище изменилось: %q", got)
		}
	})

	t.Run("изменения загруженной копии не попадают в хранилище", func(t *testing.T) {
		s := factory(t, sampleDump)

		first := mustLoad(t, s)
		deleteIDs(2)(t, first)

		if got := dump(t, mustLoad(t, s)); got != sampleDump {
			t.Errorf("хранилище изменилось: %q", got)
		}
	})

	t.Run("каждая загрузка возвращает новый инвентарь", func(t *testing.T) {
		s := factory(t, sampleDump)

		if first, second := mustLoad(t, s), mustLoad(t, s); first == second {
			t.Error("Load вернул один и тот же *Inventory дважды")
		}
	})
}

func TestMemoryStore(t *testing.T) {
	t.Run("нулевое значение работает как пустое хранилище", func(t *testing.T) {
		var s MemoryStore

		if n := len(mustLoad(t, &s).Entries()); n != 0 {
			t.Errorf("предметов = %d, хочу 0", n)
		}
		mustSave(t, &s, sampleInventory(t))
		if got := dump(t, mustLoad(t, &s)); got != sampleDump {
			t.Errorf("загружено %q", got)
		}
	})

	t.Run("NewMemoryStore пуст", func(t *testing.T) {
		if n := len(mustLoad(t, NewMemoryStore()).Entries()); n != 0 {
			t.Errorf("предметов = %d, хочу 0", n)
		}
	})

	t.Run("Save заменяет повреждённые данные", func(t *testing.T) {
		s := NewMemoryStoreFrom("бред\n")
		if _, err := s.Load(); err == nil {
			t.Fatal("повреждённые данные должны давать ошибку")
		}

		mustSave(t, s, sampleInventory(t))

		if got := dump(t, mustLoad(t, s)); got != sampleDump {
			t.Errorf("загружено %q", got)
		}
	})

	t.Run("ошибки содержат контекст", func(t *testing.T) {
		_, loadErr := NewMemoryStoreFrom("бред\n").Load()
		if loadErr == nil || !strings.Contains(loadErr.Error(), "load from memory") {
			t.Errorf("ошибка загрузки = %v, хочу контекст «load from memory»", loadErr)
		}
		saveErr := NewMemoryStore().Save(newInventory(t, fakeItem{}))
		if saveErr == nil || !strings.Contains(saveErr.Error(), "save to memory") {
			t.Errorf("ошибка сохранения = %v, хочу контекст «save to memory»", saveErr)
		}
	})
}
