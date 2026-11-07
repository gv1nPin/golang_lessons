package cli

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"golang_lessons/lesson_7/game/inventory"
	"golang_lessons/lesson_7/game/storage"
)

// fakeStore — тестовая реализация Store поверх MemoryStore: считает вызовы
// и умеет возвращать заданные ошибки загрузки и сохранения.
type fakeStore struct {
	backing *storage.MemoryStore
	loadErr error
	saveErr error
	loads   int
	saves   int
}

var _ Store = (*fakeStore)(nil)

func newFakeStore() *fakeStore {
	return &fakeStore{backing: storage.NewMemoryStore()}
}

func (s *fakeStore) Load() (*inventory.Inventory, error) {
	s.loads++
	if s.loadErr != nil {
		return nil, s.loadErr
	}
	return s.backing.Load()
}

func (s *fakeStore) Save(inv *inventory.Inventory) error {
	s.saves++
	if s.saveErr != nil {
		return s.saveErr
	}
	return s.backing.Save(inv)
}

// seededStore создаёт хранилище с предметами (ID 1, 2, ...), не считая это обращением к хранилищу.
func seededStore(t *testing.T, items ...inventory.Item) *fakeStore {
	t.Helper()
	inv := inventory.New()
	for _, item := range items {
		if _, err := inv.AddItem(item); err != nil {
			t.Fatal(err)
		}
	}
	s := newFakeStore()
	if err := s.backing.Save(inv); err != nil {
		t.Fatal(err)
	}
	return s
}

// stored возвращает предметы, которые реально лежат в хранилище.
func stored(t *testing.T, s *fakeStore) []inventory.Entry {
	t.Helper()
	inv, err := s.backing.Load()
	if err != nil {
		t.Fatal(err)
	}
	return inv.Entries()
}

func storedIDs(t *testing.T, s *fakeStore) []int {
	t.Helper()
	var ids []int
	for _, entry := range stored(t, s) {
		ids = append(ids, entry.ID)
	}
	return ids
}

func mustWeapon(t *testing.T, name string, weight float64, damage, durability int) *inventory.Weapon {
	t.Helper()
	w, err := inventory.NewWeapon(name, weight, damage, durability)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func mustArmor(t *testing.T, name string, weight float64, defense int, equipped bool) *inventory.Armor {
	t.Helper()
	a, err := inventory.NewArmor(name, weight, defense)
	if err != nil {
		t.Fatal(err)
	}
	a.Equipped = equipped
	return a
}

func mustPotion(t *testing.T, name string, weight float64, effect string, charges int) *inventory.Potion {
	t.Helper()
	p, err := inventory.NewPotion(name, weight, effect, charges)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func asWeapon(t *testing.T, item inventory.Item) *inventory.Weapon {
	t.Helper()
	w, ok := item.(*inventory.Weapon)
	if !ok {
		t.Fatalf("хочу оружие, получено %T", item)
	}
	return w
}

func asArmor(t *testing.T, item inventory.Item) *inventory.Armor {
	t.Helper()
	a, ok := item.(*inventory.Armor)
	if !ok {
		t.Fatalf("хочу броню, получено %T", item)
	}
	return a
}

func asPotion(t *testing.T, item inventory.Item) *inventory.Potion {
	t.Helper()
	p, ok := item.(*inventory.Potion)
	if !ok {
		t.Fatalf("хочу зелье, получено %T", item)
	}
	return p
}

type result struct {
	code   int
	stdout string
	stderr string
}

// runCLI выполняет одну команду над хранилищем s.
func runCLI(s Store, args ...string) result {
	var stdout, stderr bytes.Buffer
	code := New(s, &stdout, &stderr).Run(args)
	return result{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

// Неверное использование всегда даёт код 2, ничего не пишет в stdout и не обращается к хранилищу.
func TestUsageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"nil вместо аргументов", nil},
		{"пустые аргументы", []string{}},
		{"неизвестная команда", []string{"fly"}},
		{"help не является командой", []string{"help"}},
		{"-h на месте команды", []string{"-h"}},
		{"add: неизвестный флаг", []string{"add", "-bogus"}},
		{"add: нет -name", []string{"add", "-weight", "1"}},
		{"add: пустое -name", []string{"add", "-name", ""}},
		{"add: неизвестный тип", []string{"add", "-name", "x", "-type", "ring"}},
		{"add: отрицательный вес", []string{"add", "-name", "x", "-weight=-5"}},
		{"add: NaN в весе", []string{"add", "-name", "x", "-weight=NaN"}},
		{"add: вес не число", []string{"add", "-name", "x", "-weight", "abc"}},
		{"add: урон не число", []string{"add", "-name", "x", "-damage", "abc"}},
		{"add: разделитель в названии", []string{"add", "-name", "a|b"}},
		{"add: перевод строки в названии", []string{"add", "-name", "a\nb"}},
		{"add: разделитель в эффекте зелья", []string{"add", "-name", "x", "-type", "potion", "-effect", "a|b"}},
		{"add: лишний аргумент", []string{"add", "-name", "x", "extra"}},
		{"list: лишний аргумент", []string{"list", "extra"}},
		{"list: неизвестный флаг", []string{"list", "-x"}},
		{"use: нет -id", []string{"use"}},
		{"use: -id равен 0", []string{"use", "-id", "0"}},
		{"use: отрицательный -id", []string{"use", "-id=-1"}},
		{"use: -id не число", []string{"use", "-id", "abc"}},
		{"use: лишний аргумент", []string{"use", "-id", "1", "extra"}},
		{"delete: нет -id", []string{"delete"}},
		{"delete: -id равен 0", []string{"delete", "-id", "0"}},
		{"delete: отрицательный -id", []string{"delete", "-id=-7"}},
		{"delete: лишний аргумент", []string{"delete", "-id", "1", "extra"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := seededStore(t, mustWeapon(t, "Меч", 1, 1, 1))

			res := runCLI(s, tt.args...)

			if res.code != exitUsage {
				t.Errorf("код = %d, хочу %d", res.code, exitUsage)
			}
			if res.stdout != "" {
				t.Errorf("stdout = %q, хочу пустой", res.stdout)
			}
			if res.stderr == "" {
				t.Error("stderr пустой: пользователь не узнает, что пошло не так")
			}
			if s.loads != 0 || s.saves != 0 {
				t.Errorf("хранилище тронуто: Load %d раз, Save %d раз", s.loads, s.saves)
			}
			if got := storedIDs(t, s); len(got) != 1 || got[0] != 1 {
				t.Errorf("содержимое хранилища изменилось: %v", got)
			}
		})
	}
}

// Точные тексты сообщений и коды: результаты идут в stdout, ошибки в stderr.
func TestExactOutput(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{"без аргументов", nil, exitUsage, "", "usage: inventory <command> [flags]\n"},
		{"неизвестная команда", []string{"fly"}, exitUsage, "", "unknown command: fly\n"},
		{"add без имени", []string{"add"}, exitUsage, "", "flag -name is required\n"},
		{"add с неизвестным типом", []string{"add", "-name", "x", "-type", "ring"}, exitUsage, "",
			"unknown type \"ring\" (use weapon|armor|potion)\n"},
		{"add с неизвестным флагом", []string{"add", "-bogus"}, exitUsage, "", "flag provided but not defined: -bogus\n"},
		{"use без -id", []string{"use"}, exitUsage, "", "flag -id is required and must be positive\n"},
		{"delete с -id 0", []string{"delete", "-id", "0"}, exitUsage, "", "flag -id is required and must be positive\n"},
		{"list с лишним аргументом", []string{"list", "extra"}, exitUsage, "",
			"unexpected positional arguments for list: [extra]\n"},
		{"list пустого инвентаря", []string{"list"}, exitOK, "inventory is empty\n", ""},
		{"use несуществующего предмета", []string{"use", "-id", "42"}, exitFailure, "", "item id=42 not found\n"},
		{"delete несуществующего предмета", []string{"delete", "-id", "42"}, exitFailure, "", "item id=42 not found\n"},
		{"add сообщает ID", []string{"add", "-name", "Меч"}, exitOK, "added item id=1\n", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := runCLI(newFakeStore(), tt.args...)

			if res.code != tt.wantCode {
				t.Errorf("код = %d, хочу %d", res.code, tt.wantCode)
			}
			if res.stdout != tt.wantStdout {
				t.Errorf("stdout = %q, хочу %q", res.stdout, tt.wantStdout)
			}
			if res.stderr != tt.wantStderr {
				t.Errorf("stderr = %q, хочу %q", res.stderr, tt.wantStderr)
			}
		})
	}
}

func TestList(t *testing.T) {
	s := seededStore(t,
		mustWeapon(t, "Меч", 2.5, 10, 5),
		mustArmor(t, "Щит", 4.5, 5, false),
		mustPotion(t, "Лечебное", 0.3, "+50 HP", 3),
	)

	t.Run("все предметы с ID", func(t *testing.T) {
		res := runCLI(s, "list")

		want := "id=1  Меч (вес: 2.5)\nid=2  Щит (вес: 4.5)\nid=3  Лечебное (вес: 0.3)\n"
		if res.code != exitOK || res.stdout != want || res.stderr != "" {
			t.Errorf("код %d, stdout %q, stderr %q; хочу stdout %q", res.code, res.stdout, res.stderr, want)
		}
		if s.loads != 1 || s.saves != 0 {
			t.Errorf("Load %d раз, Save %d раз; list только читает", s.loads, s.saves)
		}
	})

	t.Run("ID не перенумеровываются после удаления", func(t *testing.T) {
		if res := runCLI(s, "delete", "-id", "2"); res.code != exitOK {
			t.Fatalf("delete: код %d, stderr %q", res.code, res.stderr)
		}

		res := runCLI(s, "list")

		want := "id=1  Меч (вес: 2.5)\nid=3  Лечебное (вес: 0.3)\n"
		if res.stdout != want {
			t.Errorf("stdout = %q, хочу %q", res.stdout, want)
		}
	})
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		check func(t *testing.T, item inventory.Item)
	}{
		{"оружие по умолчанию", []string{"add", "-name", "Меч"}, func(t *testing.T, item inventory.Item) {
			w := asWeapon(t, item)
			if w.Name() != "Меч" || w.Weight() != 0 || w.Damage != 5 || w.Durability != 10 {
				t.Errorf("создано %+v", *w)
			}
		}},
		{"оружие со своими значениями", []string{"add", "-name", "Топор", "-weight", "2.5", "-damage", "7", "-durability", "3"},
			func(t *testing.T, item inventory.Item) {
				w := asWeapon(t, item)
				if w.Name() != "Топор" || w.Weight() != 2.5 || w.Damage != 7 || w.Durability != 3 {
					t.Errorf("создано %+v", *w)
				}
			}},
		{"название с пробелами", []string{"add", "-name", "Старый меч"}, func(t *testing.T, item inventory.Item) {
			if got := asWeapon(t, item).Name(); got != "Старый меч" {
				t.Errorf("название = %q", got)
			}
		}},
		{"броня по умолчанию", []string{"add", "-name", "Щит", "-type", "armor"}, func(t *testing.T, item inventory.Item) {
			a := asArmor(t, item)
			if a.Name() != "Щит" || a.Defense != 5 || a.Equipped {
				t.Errorf("создано %+v", *a)
			}
		}},
		{"броня со своими значениями", []string{"add", "-name", "Латы", "-type", "armor", "-defense", "9", "-weight", "12"},
			func(t *testing.T, item inventory.Item) {
				a := asArmor(t, item)
				if a.Defense != 9 || a.Weight() != 12 || a.Equipped {
					t.Errorf("создано %+v", *a)
				}
			}},
		{"зелье по умолчанию", []string{"add", "-name", "Зелье", "-type", "potion"}, func(t *testing.T, item inventory.Item) {
			p := asPotion(t, item)
			if p.Effect != "+10 HP" || p.Charges != 3 {
				t.Errorf("создано %+v", *p)
			}
		}},
		{"зелье со своими значениями", []string{"add", "-name", "Яд", "-type", "potion", "-effect", "-5 HP", "-charges", "1"},
			func(t *testing.T, item inventory.Item) {
				p := asPotion(t, item)
				if p.Effect != "-5 HP" || p.Charges != 1 {
					t.Errorf("создано %+v", *p)
				}
			}},
		{"тип без учёта регистра", []string{"add", "-name", "Зелье", "-type", "POTION"}, func(t *testing.T, item inventory.Item) {
			asPotion(t, item)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newFakeStore()

			res := runCLI(s, tt.args...)

			if res.code != exitOK || res.stdout != "added item id=1\n" || res.stderr != "" {
				t.Fatalf("код %d, stdout %q, stderr %q", res.code, res.stdout, res.stderr)
			}
			if s.saves != 1 {
				t.Errorf("Save вызван %d раз, хочу 1", s.saves)
			}
			entries := stored(t, s)
			if len(entries) != 1 || entries[0].ID != 1 {
				t.Fatalf("в хранилище %v, хочу один предмет с ID 1", entries)
			}
			tt.check(t, entries[0].Item)
		})
	}
}

// Новый ID — это наибольший существующий плюс один (файл не хранит счётчик). Поэтому внутри одного
// запуска ID не повторяются, а после удаления последнего предмета и следующего запуска его номер освобождается.
func TestAddAssignsIDsAcrossRuns(t *testing.T) {
	s := newFakeStore()

	steps := []struct {
		args       []string
		wantStdout string
		wantIDs    []int
	}{
		{[]string{"add", "-name", "Меч"}, "added item id=1\n", []int{1}},
		{[]string{"add", "-name", "Лук"}, "added item id=2\n", []int{1, 2}},
		{[]string{"delete", "-id", "1"}, "deleted item id=1\n", []int{2}},
		{[]string{"add", "-name", "Копьё"}, "added item id=3\n", []int{2, 3}},
		{[]string{"delete", "-id", "3"}, "deleted item id=3\n", []int{2}},
		{[]string{"add", "-name", "Нож"}, "added item id=3\n", []int{2, 3}},
	}
	for i, step := range steps {
		res := runCLI(s, step.args...)

		if res.code != exitOK || res.stdout != step.wantStdout {
			t.Fatalf("шаг %d %v: код %d, stdout %q, хочу %q", i, step.args, res.code, res.stdout, step.wantStdout)
		}
		got := storedIDs(t, s)
		if fmt.Sprint(got) != fmt.Sprint(step.wantIDs) {
			t.Fatalf("шаг %d: ID в хранилище %v, хочу %v", i, got, step.wantIDs)
		}
	}
}

func TestUse(t *testing.T) {
	tests := []struct {
		name       string
		item       inventory.Item
		wantCode   int
		wantStdout string // подстрока stdout при успехе
		wantStderr string // подстрока stderr при ошибке
		wantSaves  int
		check      func(t *testing.T, item inventory.Item)
	}{
		{
			name: "оружие изнашивается", item: mustWeapon(t, "Меч", 1, 10, 2),
			wantCode: exitOK, wantStdout: "прочность осталась 1", wantSaves: 1,
			check: func(t *testing.T, item inventory.Item) {
				if got := asWeapon(t, item).Durability; got != 1 {
					t.Errorf("прочность = %d, хочу 1", got)
				}
			},
		},
		{
			name: "последнее использование оружия", item: mustWeapon(t, "Меч", 1, 10, 1),
			wantCode: exitOK, wantStdout: "прочность осталась 0", wantSaves: 1,
			check: func(t *testing.T, item inventory.Item) {
				if got := asWeapon(t, item).Durability; got != 0 {
					t.Errorf("прочность = %d, хочу 0", got)
				}
			},
		},
		{
			name: "сломанное оружие", item: mustWeapon(t, "Меч", 1, 10, 0),
			wantCode: exitFailure, wantStderr: "cannot use: item is broken", wantSaves: 0,
			check: func(t *testing.T, item inventory.Item) {
				if got := asWeapon(t, item).Durability; got != 0 {
					t.Errorf("прочность = %d, хочу 0", got)
				}
			},
		},
		{
			name: "заряды зелья", item: mustPotion(t, "Лечебное", 1, "+50 HP", 3),
			wantCode: exitOK, wantStdout: "зарядов осталось 2", wantSaves: 1,
			check: func(t *testing.T, item inventory.Item) {
				if got := asPotion(t, item).Charges; got != 2 {
					t.Errorf("заряды = %d, хочу 2", got)
				}
			},
		},
		{
			name: "пустое зелье", item: mustPotion(t, "Лечебное", 1, "+50 HP", 0),
			wantCode: exitFailure, wantStderr: "cannot use: item is exhausted", wantSaves: 0,
			check: func(t *testing.T, item inventory.Item) {
				if got := asPotion(t, item).Charges; got != 0 {
					t.Errorf("заряды = %d, хочу 0", got)
				}
			},
		},
		{
			name: "броня надевается", item: mustArmor(t, "Щит", 4.5, 5, false),
			wantCode: exitOK, wantStdout: "Надевается Щит", wantSaves: 1,
			check: func(t *testing.T, item inventory.Item) {
				if !asArmor(t, item).Equipped {
					t.Error("броня должна быть надета")
				}
			},
		},
		{
			name: "уже надетая броня", item: mustArmor(t, "Щит", 4.5, 5, true),
			wantCode: exitOK, wantStdout: "уже надето", wantSaves: 1,
			check: func(t *testing.T, item inventory.Item) {
				if !asArmor(t, item).Equipped {
					t.Error("броня должна остаться надетой")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := seededStore(t, tt.item)

			res := runCLI(s, "use", "-id", "1")

			if res.code != tt.wantCode {
				t.Fatalf("код = %d, хочу %d (stderr %q)", res.code, tt.wantCode, res.stderr)
			}
			if tt.wantCode == exitOK {
				if !strings.Contains(res.stdout, tt.wantStdout) || res.stderr != "" {
					t.Errorf("stdout %q (хочу подстроку %q), stderr %q", res.stdout, tt.wantStdout, res.stderr)
				}
			} else if res.stdout != "" || !strings.Contains(res.stderr, tt.wantStderr) {
				t.Errorf("stdout %q, stderr %q (хочу подстроку %q)", res.stdout, res.stderr, tt.wantStderr)
			}
			if s.saves != tt.wantSaves {
				t.Errorf("Save вызван %d раз, хочу %d", s.saves, tt.wantSaves)
			}
			entries := stored(t, s)
			if len(entries) != 1 {
				t.Fatalf("в хранилище %d предметов, хочу 1", len(entries))
			}
			tt.check(t, entries[0].Item)
		})
	}
}

// Состояние накапливается между запусками команд: каждый запуск видит результат предыдущего.
func TestUseAcrossRuns(t *testing.T) {
	s := seededStore(t, mustWeapon(t, "Меч", 1, 10, 2))

	steps := []struct {
		wantCode int
		wantText string
	}{
		{exitOK, "прочность осталась 1"},
		{exitOK, "прочность осталась 0"},
		{exitFailure, "item is broken"},
		{exitFailure, "item is broken"},
	}
	for i, step := range steps {
		res := runCLI(s, "use", "-id", "1")

		if res.code != step.wantCode || !strings.Contains(res.stdout+res.stderr, step.wantText) {
			t.Fatalf("запуск %d: код %d, stdout %q, stderr %q; хочу код %d и %q",
				i+1, res.code, res.stdout, res.stderr, step.wantCode, step.wantText)
		}
	}
	if s.saves != 2 {
		t.Errorf("Save вызван %d раз, хочу 2: только за успешные использования", s.saves)
	}
}

func TestDelete(t *testing.T) {
	s := seededStore(t,
		mustWeapon(t, "Меч", 1, 1, 1),
		mustArmor(t, "Щит", 1, 1, false),
		mustPotion(t, "Зелье", 1, "x", 1),
	)

	res := runCLI(s, "delete", "-id", "2")

	if res.code != exitOK || res.stdout != "deleted item id=2\n" || res.stderr != "" {
		t.Fatalf("код %d, stdout %q, stderr %q", res.code, res.stdout, res.stderr)
	}
	if got := storedIDs(t, s); fmt.Sprint(got) != "[1 3]" {
		t.Errorf("ID в хранилище %v, хочу [1 3]", got)
	}
	if s.saves != 1 {
		t.Errorf("Save вызван %d раз, хочу 1", s.saves)
	}

	t.Run("повторное удаление не сохраняет", func(t *testing.T) {
		res := runCLI(s, "delete", "-id", "2")

		if res.code != exitFailure || res.stderr != "item id=2 not found\n" {
			t.Errorf("код %d, stderr %q", res.code, res.stderr)
		}
		if s.saves != 1 {
			t.Errorf("Save вызван %d раз, хочу по-прежнему 1", s.saves)
		}
	})
}

func TestStoreLoadFailure(t *testing.T) {
	commands := []struct {
		name string
		args []string
	}{
		{"list", []string{"list"}},
		{"add", []string{"add", "-name", "Меч"}},
		{"use", []string{"use", "-id", "1"}},
		{"delete", []string{"delete", "-id", "1"}},
	}
	for _, tt := range commands {
		t.Run(tt.name, func(t *testing.T) {
			s := seededStore(t, mustWeapon(t, "Меч", 1, 1, 2))
			s.loadErr = errors.New("disk read failed")

			res := runCLI(s, tt.args...)

			if res.code != exitFailure {
				t.Errorf("код = %d, хочу %d", res.code, exitFailure)
			}
			if res.stdout != "" {
				t.Errorf("stdout = %q, хочу пустой", res.stdout)
			}
			if res.stderr != "load inventory: disk read failed\n" {
				t.Errorf("stderr = %q", res.stderr)
			}
			if s.saves != 0 {
				t.Errorf("после неудачной загрузки Save вызван %d раз", s.saves)
			}
		})
	}
}

func TestStoreSaveFailure(t *testing.T) {
	commands := []struct {
		name string
		args []string
	}{
		{"add", []string{"add", "-name", "Лук"}},
		{"use", []string{"use", "-id", "1"}},
		{"delete", []string{"delete", "-id", "1"}},
	}
	for _, tt := range commands {
		t.Run(tt.name, func(t *testing.T) {
			s := seededStore(t, mustWeapon(t, "Меч", 1, 1, 2))
			s.saveErr = errors.New("disk full")

			res := runCLI(s, tt.args...)

			if res.code != exitFailure {
				t.Errorf("код = %d, хочу %d", res.code, exitFailure)
			}
			if res.stdout != "" {
				t.Errorf("об успехе нельзя сообщать, если сохранение не удалось: stdout = %q", res.stdout)
			}
			if res.stderr != "save: disk full\n" {
				t.Errorf("stderr = %q", res.stderr)
			}

			// Хранилище осталось прежним: один меч с прочностью 2.
			entries := stored(t, s)
			if len(entries) != 1 || entries[0].ID != 1 {
				t.Fatalf("содержимое хранилища изменилось: %v", entries)
			}
			if got := asWeapon(t, entries[0].Item).Durability; got != 2 {
				t.Errorf("прочность в хранилище = %d, хочу 2", got)
			}
		})
	}
}

// jammedItem — предмет с неожиданной ошибкой Use. Среди настоящих предметов такого нет,
// но его можно подставить через заглушку хранилища и проверить запасную ветку команды use.
type jammedItem struct{}

func (jammedItem) Name() string         { return "Заклинивший" }
func (jammedItem) Weight() float64      { return 1 }
func (jammedItem) Use() (string, error) { return "", errors.New("jammed") }

// staticStore отдаёт готовый инвентарь без кодирования и считает сохранения.
type staticStore struct {
	inv   *inventory.Inventory
	saves int
}

func (s *staticStore) Load() (*inventory.Inventory, error) { return s.inv, nil }
func (s *staticStore) Save(*inventory.Inventory) error     { s.saves++; return nil }

func TestUseUnexpectedError(t *testing.T) {
	inv := inventory.New()
	if _, err := inv.AddItem(jammedItem{}); err != nil {
		t.Fatal(err)
	}
	s := &staticStore{inv: inv}

	res := runCLI(s, "use", "-id", "1")

	if res.code != exitFailure || res.stdout != "" || res.stderr != "use error: jammed\n" {
		t.Errorf("код %d, stdout %q, stderr %q", res.code, res.stdout, res.stderr)
	}
	if s.saves != 0 {
		t.Errorf("Save вызван %d раз, хочу 0: использование не удалось", s.saves)
	}
}

func TestLoadErrorDetails(t *testing.T) {
	t.Run("повреждённая строка: номер и содержимое через errors.As", func(t *testing.T) {
		s := newFakeStore()
		parseErr := &storage.ParseItemError{Line: 7, Value: "bad|line", Err: errors.New("boom")}
		s.loadErr = fmt.Errorf("load /some/path: %w", parseErr)

		res := runCLI(s, "list")

		if res.code != exitFailure || res.stdout != "" {
			t.Errorf("код %d, stdout %q", res.code, res.stdout)
		}
		if want := `line 7: "bad|line"`; !strings.Contains(res.stderr, want) {
			t.Errorf("stderr %q не содержит %q", res.stderr, want)
		}
	})

	t.Run("обычная ошибка: без номера строки", func(t *testing.T) {
		s := newFakeStore()
		s.loadErr = errors.New("permission denied")

		res := runCLI(s, "list")

		if strings.Contains(res.stderr, "line") {
			t.Errorf("stderr %q не должен упоминать строку", res.stderr)
		}
	})

	t.Run("настоящее хранилище с повреждёнными данными", func(t *testing.T) {
		s := storage.NewMemoryStoreFrom("Weapon|1|Меч|10|5|3.5\nмусор\n")

		res := runCLI(s, "list")

		if res.code != exitFailure || !strings.Contains(res.stderr, `line 2: "мусор"`) {
			t.Errorf("код %d, stderr %q", res.code, res.stderr)
		}
		// Неверное использование распознаётся раньше обращения к хранилищу.
		if res := runCLI(s, "fly"); res.code != exitUsage {
			t.Errorf("неизвестная команда: код %d, хочу %d", res.code, exitUsage)
		}
		if res := runCLI(s, "add", "-bogus"); res.code != exitUsage {
			t.Errorf("неверный флаг: код %d, хочу %d", res.code, exitUsage)
		}
	})
}

func TestHelp(t *testing.T) {
	tests := []struct {
		command string
		want    []string // подстроки справки по флагам
	}{
		{"list", nil},
		{"add", []string{"-name", "-weight", "-type", "-damage", "-durability", "-defense", "-effect", "-charges"}},
		{"use", []string{"-id"}},
		{"delete", []string{"-id"}},
	}
	for _, tt := range tests {
		for _, flag := range []string{"-h", "-help"} {
			t.Run(tt.command+" "+flag, func(t *testing.T) {
				s := newFakeStore()

				res := runCLI(s, tt.command, flag)

				if res.code != exitOK || res.stderr != "" {
					t.Errorf("код %d, stderr %q; справка не ошибка", res.code, res.stderr)
				}
				for _, want := range tt.want {
					if !strings.Contains(res.stdout, want) {
						t.Errorf("в справке нет %q: %q", want, res.stdout)
					}
				}
				if s.loads != 0 || s.saves != 0 {
					t.Errorf("справка не должна трогать хранилище: Load %d, Save %d", s.loads, s.saves)
				}
			})
		}
	}
}
