package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"golang_lessons/lesson_6/game/inventory"
	"golang_lessons/lesson_6/game/storage"
)

// failingStore — заглушка: загрузка работает, сохранение всегда падает.
// Подставляется вместо настоящего хранилища благодаря интерфейсу Store.
type failingStore struct{}

func (failingStore) Load() (*inventory.Inventory, error) { return inventory.New(), nil }
func (failingStore) Save(*inventory.Inventory) error     { return errors.New("disk full") }

// runCLI выполняет одну команду и возвращает код завершения, stdout и stderr.
func runCLI(store Store, args ...string) (code int, stdout, stderr string) {
	var out, errOut bytes.Buffer
	code = New(store, &out, &errOut).Run(args)
	return code, out.String(), errOut.String()
}

func TestExitCodes(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"без аргументов", nil, exitUsage},
		{"неизвестная команда", []string{"fly"}, exitUsage},
		{"неверный флаг", []string{"add", "-bogus"}, exitUsage},
		{"add без -name", []string{"add", "-weight", "1"}, exitUsage},
		{"add с неизвестным типом", []string{"add", "-name", "x", "-type", "ring"}, exitUsage},
		{"add с отрицательным весом", []string{"add", "-name", "x", "-weight=-5"}, exitUsage},
		{"add с разделителем в названии", []string{"add", "-name", "a|b"}, exitUsage},
		{"лишний аргумент у list", []string{"list", "extra"}, exitUsage},
		{"лишний аргумент у add", []string{"add", "-name", "x", "extra"}, exitUsage},
		{"use без -id", []string{"use"}, exitUsage},
		{"use с отрицательным -id", []string{"use", "-id=-1"}, exitUsage},
		{"delete без -id", []string{"delete"}, exitUsage},
		{"list пустого инвентаря", []string{"list"}, exitOK},
		{"use несуществующего предмета", []string{"use", "-id", "42"}, exitFailure},
		{"delete несуществующего предмета", []string{"delete", "-id", "42"}, exitFailure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, _ := runCLI(storage.NewMemoryStore(), tt.args...)
			if code != tt.want {
				t.Errorf("код = %d, хочу %d", code, tt.want)
			}
		})
	}
}

func TestAddUseDeleteScenario(t *testing.T) {
	store := storage.NewMemoryStore()

	code, stdout, stderr := runCLI(store, "add", "-name", "Зелье", "-type", "potion", "-charges", "1")
	if code != exitOK || !strings.Contains(stdout, "id=1") || stderr != "" {
		t.Fatalf("add: код %d, stdout %q, stderr %q", code, stdout, stderr)
	}

	code, stdout, stderr = runCLI(store, "use", "-id", "1")
	if code != exitOK || !strings.Contains(stdout, "Выпито") || stderr != "" {
		t.Fatalf("use: код %d, stdout %q, stderr %q", code, stdout, stderr)
	}

	// Заряды кончились: ошибка идёт в stderr, в stdout ничего нет.
	code, stdout, stderr = runCLI(store, "use", "-id", "1")
	if code != exitFailure || stdout != "" || !strings.Contains(stderr, "exhausted") {
		t.Fatalf("use пустого зелья: код %d, stdout %q, stderr %q", code, stdout, stderr)
	}

	code, stdout, _ = runCLI(store, "list")
	if code != exitOK || !strings.Contains(stdout, "id=1  Зелье") {
		t.Fatalf("list: код %d, stdout %q", code, stdout)
	}

	code, stdout, _ = runCLI(store, "delete", "-id", "1")
	if code != exitOK || !strings.Contains(stdout, "id=1") {
		t.Fatalf("delete: код %d, stdout %q", code, stdout)
	}

	code, _, stderr = runCLI(store, "delete", "-id", "1")
	if code != exitFailure || !strings.Contains(stderr, "not found") {
		t.Fatalf("повторный delete: код %d, stderr %q", code, stderr)
	}

	_, stdout, _ = runCLI(store, "list")
	if !strings.Contains(stdout, "empty") {
		t.Errorf("после удаления list = %q, хочу пустой инвентарь", stdout)
	}
}

func TestUseChangesAndPersistsState(t *testing.T) {
	store := storage.NewMemoryStore()
	runCLI(store, "add", "-name", "Меч", "-durability", "1")
	runCLI(store, "add", "-name", "Щит", "-type", "armor")

	if code, _, _ := runCLI(store, "use", "-id", "1"); code != exitOK {
		t.Fatalf("первое использование меча: код %d", code)
	}
	code, _, stderr := runCLI(store, "use", "-id", "1")
	if code != exitFailure || !strings.Contains(stderr, "broken") {
		t.Errorf("сломанный меч: код %d, stderr %q", code, stderr)
	}

	_, stdout, _ := runCLI(store, "use", "-id", "2")
	if !strings.Contains(stdout, "Надевается") {
		t.Errorf("первое использование брони: %q", stdout)
	}
	_, stdout, _ = runCLI(store, "use", "-id", "2")
	if !strings.Contains(stdout, "уже надето") {
		t.Errorf("повторное использование брони: %q", stdout)
	}

	// Состояние действительно сохранилось в хранилище.
	inv, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	item, err := inv.Get(2)
	if err != nil {
		t.Fatal(err)
	}
	if armor, ok := item.(*inventory.Armor); !ok || !armor.Equipped {
		t.Errorf("броня после use = %#v, хочу надетую", item)
	}
}

func TestInvalidInputDoesNotChangeStore(t *testing.T) {
	store := storage.NewMemoryStore()

	if code, _, _ := runCLI(store, "add", "-name", "a|b"); code != exitUsage {
		t.Fatalf("код = %d, хочу %d", code, exitUsage)
	}
	_, stdout, _ := runCLI(store, "list")
	if !strings.Contains(stdout, "empty") {
		t.Errorf("list = %q, хочу пустой инвентарь", stdout)
	}
}

func TestCorruptedStore(t *testing.T) {
	store := storage.NewMemoryStoreFrom("мусор\n")

	code, stdout, stderr := runCLI(store, "list")
	if code != exitFailure || stdout != "" || !strings.Contains(stderr, "line 1") {
		t.Errorf("list: код %d, stdout %q, stderr %q", code, stdout, stderr)
	}

	// Неверное использование распознаётся раньше обращения к хранилищу.
	if code, _, _ := runCLI(store, "fly"); code != exitUsage {
		t.Errorf("неизвестная команда: код %d, хочу %d", code, exitUsage)
	}
	if code, _, _ := runCLI(store, "add", "-bogus"); code != exitUsage {
		t.Errorf("неверный флаг: код %d, хочу %d", code, exitUsage)
	}
}

func TestSaveFailure(t *testing.T) {
	code, stdout, stderr := runCLI(failingStore{}, "add", "-name", "Меч")
	if code != exitFailure || stdout != "" || !strings.Contains(stderr, "save") {
		t.Errorf("код %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestHelpFlag(t *testing.T) {
	code, stdout, _ := runCLI(storage.NewMemoryStore(), "add", "-h")
	if code != exitOK || !strings.Contains(stdout, "-name") {
		t.Errorf("код %d, stdout %q", code, stdout)
	}
}
