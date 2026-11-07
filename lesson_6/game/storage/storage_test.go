package storage

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"golang_lessons/lesson_6/game/inventory"
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

// sampleDump — как sampleInventory выглядит в формате хранения.
const sampleDump = "Weapon|1|Меч|10|5|3.5\nArmor|2|Щит|5|4.5|1\nPotion|3|Лечебное|+50 HP|3|0.3\n"

func sampleInventory(t *testing.T) *inventory.Inventory {
	t.Helper()
	sword, err := inventory.NewWeapon("Меч", 3.5, 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	shield, err := inventory.NewArmor("Щит", 4.5, 5)
	if err != nil {
		t.Fatal(err)
	}
	shield.Equipped = true
	potion, err := inventory.NewPotion("Лечебное", 0.3, "+50 HP", 3)
	if err != nil {
		t.Fatal(err)
	}

	inv := inventory.New()
	for _, item := range []inventory.Item{sword, shield, potion} {
		if _, err := inv.AddItem(item); err != nil {
			t.Fatal(err)
		}
	}
	return inv
}

// dump превращает инвентарь в текст формата хранения: так удобно сравнивать содержимое.
func dump(t *testing.T, inv *inventory.Inventory) string {
	t.Helper()
	var buf bytes.Buffer
	if err := encode(&buf, inv); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

// fakeItem — предмет неизвестного хранилищу типа.
type fakeItem struct{}

func (fakeItem) Name() string         { return "Пустышка" }
func (fakeItem) Weight() float64      { return 1 }
func (fakeItem) Use() (string, error) { return "", nil }

// TestStoreContract запускает одни и те же проверки для обоих хранилищ: правила сохранения,
// загрузки и обработки ошибок у них должны совпадать.
func TestStoreContract(t *testing.T) {
	factories := map[string]storeFactory{
		"file":   fileFactory,
		"memory": memoryFactory,
	}

	for name, factory := range factories {
		t.Run(name, func(t *testing.T) {
			t.Run("пустые данные дают пустой инвентарь", func(t *testing.T) {
				inv, err := factory(t, "").Load()
				if err != nil {
					t.Fatal(err)
				}
				if n := len(inv.Entries()); n != 0 {
					t.Errorf("предметов = %d, хочу 0", n)
				}
			})

			t.Run("сохранённое загружается без изменений", func(t *testing.T) {
				s := factory(t, "")
				if err := s.Save(sampleInventory(t)); err != nil {
					t.Fatal(err)
				}
				loaded, err := s.Load()
				if err != nil {
					t.Fatal(err)
				}
				if got := dump(t, loaded); got != sampleDump {
					t.Errorf("загружено:\n%s\nхочу:\n%s", got, sampleDump)
				}

				sword, err := inventory.NewWeapon("Копьё", 2, 1, 1)
				if err != nil {
					t.Fatal(err)
				}
				if id, err := loaded.AddItem(sword); err != nil || id != 4 {
					t.Errorf("следующий ID = %d (err %v), хочу 4", id, err)
				}
			})

			t.Run("новое содержимое короче старого", func(t *testing.T) {
				s := factory(t, "")
				inv := sampleInventory(t)
				if err := s.Save(inv); err != nil {
					t.Fatal(err)
				}
				if err := inv.Delete(2); err != nil {
					t.Fatal(err)
				}
				if err := inv.Delete(3); err != nil {
					t.Fatal(err)
				}
				if err := s.Save(inv); err != nil {
					t.Fatal(err)
				}

				loaded, err := s.Load()
				if err != nil {
					t.Fatal(err)
				}
				if got, want := dump(t, loaded), "Weapon|1|Меч|10|5|3.5\n"; got != want {
					t.Errorf("загружено %q, хочу %q", got, want)
				}
			})

			t.Run("повреждённая строка даёт ParseItemError с номером и содержимым", func(t *testing.T) {
				content := "Weapon|1|Меч|10|5|3.5\nArmor|2|Щит|abc|4.5|0\nPotion|3|Лечебное|+50 HP|3|0.3\n"
				_, err := factory(t, content).Load()

				var parseErr *ParseItemError
				if !errors.As(err, &parseErr) {
					t.Fatalf("err = %v, хочу *ParseItemError", err)
				}
				if parseErr.Line != 2 || parseErr.Value != "Armor|2|Щит|abc|4.5|0" {
					t.Errorf("Line = %d, Value = %q", parseErr.Line, parseErr.Value)
				}
			})

			t.Run("повторяющийся ID считается повреждением", func(t *testing.T) {
				content := "Weapon|1|Меч|10|5|3.5\nWeapon|1|Лук|5|0|1.2\n"
				_, err := factory(t, content).Load()

				var parseErr *ParseItemError
				if !errors.As(err, &parseErr) || parseErr.Line != 2 {
					t.Fatalf("err = %v, хочу *ParseItemError для строки 2", err)
				}
				if !errors.Is(err, inventory.ErrDuplicateID) {
					t.Errorf("err = %v, хочу ErrDuplicateID", err)
				}
			})

			t.Run("неверные значения отклоняются теми же правилами, что и в конструкторах", func(t *testing.T) {
				_, err := factory(t, "Weapon|1|Меч|10|5|-1\n").Load()
				if !errors.Is(err, inventory.ErrInvalidItem) {
					t.Errorf("err = %v, хочу ErrInvalidItem", err)
				}
			})

			t.Run("старая запись брони без Equipped читается", func(t *testing.T) {
				inv, err := factory(t, "Armor|1|Щит|5|4.5\n").Load()
				if err != nil {
					t.Fatal(err)
				}
				item, err := inv.Get(1)
				if err != nil {
					t.Fatal(err)
				}
				armor, ok := item.(*inventory.Armor)
				if !ok || armor.Equipped {
					t.Errorf("item = %#v, хочу ненадетую броню", item)
				}
			})

			t.Run("ошибка сохранения не портит прежние данные", func(t *testing.T) {
				s := factory(t, "")
				if err := s.Save(sampleInventory(t)); err != nil {
					t.Fatal(err)
				}

				bad := inventory.New()
				if _, err := bad.AddItem(fakeItem{}); err != nil {
					t.Fatal(err)
				}
				if err := s.Save(bad); !errors.Is(err, ErrNotStorable) {
					t.Fatalf("err = %v, хочу ErrNotStorable", err)
				}

				loaded, err := s.Load()
				if err != nil {
					t.Fatal(err)
				}
				if got := dump(t, loaded); got != sampleDump {
					t.Errorf("после неудачного Save данные изменились:\n%s", got)
				}
			})

			t.Run("Load и Save не делят состояние с вызывающим кодом", func(t *testing.T) {
				s := factory(t, "")
				inv := sampleInventory(t)
				if err := s.Save(inv); err != nil {
					t.Fatal(err)
				}

				// Изменения инвентаря после Save не попадают в хранилище.
				if err := inv.Delete(1); err != nil {
					t.Fatal(err)
				}
				// Изменения загруженной копии не попадают в хранилище.
				first, err := s.Load()
				if err != nil {
					t.Fatal(err)
				}
				if err := first.Delete(2); err != nil {
					t.Fatal(err)
				}

				second, err := s.Load()
				if err != nil {
					t.Fatal(err)
				}
				if got := dump(t, second); got != sampleDump {
					t.Errorf("хранилище изменилось:\n%s", got)
				}
			})
		})
	}
}

func TestFileStoreMissingFile(t *testing.T) {
	s := NewFileStore(filepath.Join(t.TempDir(), "нет.dat"))

	inv, err := s.Load()
	if err != nil {
		t.Fatalf("файла нет: err = %v, хочу nil", err)
	}
	if n := len(inv.Entries()); n != 0 {
		t.Errorf("предметов = %d, хочу 0", n)
	}
}

func TestFileStoreLeavesNoTempFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inventory.dat")
	if err := NewFileStore(path).Save(sampleInventory(t)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".tmp"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("временный файл остался: %v", err)
	}
}

// Бонус: после ошибки посередине записи прежний файл не повреждён.
func TestWriteFileAtomicKeepsOldFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inventory.dat")
	old := []byte("старое содержимое\n")
	if err := os.WriteFile(path, old, 0o644); err != nil {
		t.Fatal(err)
	}

	boom := errors.New("диск отвалился")
	err := writeFileAtomic(path, func(w io.Writer) error {
		// Больше буфера bufio (4096 байт), чтобы часть данных реально попала во временный файл.
		if _, err := io.WriteString(w, strings.Repeat("x", 10000)); err != nil {
			return err
		}
		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, хочу исходную ошибку", err)
	}

	got, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !bytes.Equal(got, old) {
		t.Errorf("прежний файл изменился: %q", got)
	}
	if _, statErr := os.Stat(path + ".tmp"); !errors.Is(statErr, fs.ErrNotExist) {
		t.Errorf("временный файл не удалён: %v", statErr)
	}
}

func TestFileStorePermissionErrors(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("права в стиле Unix недоступны или не действуют для root")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "inventory.dat")
	if err := os.WriteFile(path, []byte(sampleDump), 0o644); err != nil {
		t.Fatal(err)
	}
	s := NewFileStore(path)

	// Нет прав на чтение.
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })
	if _, err := s.Load(); !errors.Is(err, fs.ErrPermission) {
		t.Errorf("чтение: err = %v, хочу fs.ErrPermission", err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}

	// Нет прав на запись в папку: временный файл не создать.
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	if err := s.Save(sampleInventory(t)); !errors.Is(err, fs.ErrPermission) {
		t.Errorf("запись: err = %v, хочу fs.ErrPermission", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != sampleDump {
		t.Errorf("файл изменился после неудачной записи: %q", got)
	}
}
