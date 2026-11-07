package storage

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"golang_lessons/lesson_7/game/inventory"
)

// Бенчмарки сравнивают две реализации одного интерфейса store на одинаковых данных.
// Запуск: go test -bench=. -benchmem ./storage
//
// MemoryStore проходит через тот же encode и decode, что и FileStore, поэтому разница между ними —
// это цена работы с файлом: создание временного файла, запись, Sync и переименование при Save,
// открытие и чтение при Load. Отдельные BenchmarkEncode и BenchmarkDecode показывают долю самого формата.

var benchSizes = []int{10, 100, 1000}

var benchStores = []struct {
	name   string
	create func(b *testing.B) store
}{
	{"memory", func(b *testing.B) store { return NewMemoryStore() }},
	{"file", func(b *testing.B) store {
		return NewFileStore(filepath.Join(b.TempDir(), "inventory.dat"))
	}},
}

// benchInventory создаёт инвентарь из n предметов трёх типов по кругу.
func benchInventory(b *testing.B, n int) *inventory.Inventory {
	b.Helper()
	inv := inventory.New()
	for i := 0; i < n; i++ {
		var item inventory.Item
		switch i % 3 {
		case 0:
			w, err := inventory.NewWeapon(fmt.Sprintf("Меч %d", i), 3.5, 10, 5)
			if err != nil {
				b.Fatal(err)
			}
			item = w
		case 1:
			a, err := inventory.NewArmor(fmt.Sprintf("Щит %d", i), 4.5, 5)
			if err != nil {
				b.Fatal(err)
			}
			a.Equipped = i%2 == 0
			item = a
		default:
			p, err := inventory.NewPotion(fmt.Sprintf("Зелье %d", i), 0.3, "+50 HP", 3)
			if err != nil {
				b.Fatal(err)
			}
			item = p
		}
		if _, err := inv.AddItem(item); err != nil {
			b.Fatal(err)
		}
	}
	return inv
}

func BenchmarkSave(b *testing.B) {
	for _, n := range benchSizes {
		inv := benchInventory(b, n)
		for _, impl := range benchStores {
			b.Run(fmt.Sprintf("%s/items=%d", impl.name, n), func(b *testing.B) {
				s := impl.create(b)
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if err := s.Save(inv); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func BenchmarkLoad(b *testing.B) {
	for _, n := range benchSizes {
		inv := benchInventory(b, n)
		for _, impl := range benchStores {
			b.Run(fmt.Sprintf("%s/items=%d", impl.name, n), func(b *testing.B) {
				s := impl.create(b)
				if err := s.Save(inv); err != nil {
					b.Fatal(err)
				}
				// Один раз проверяем, что загружается то, что сохранено: бенчмарк не должен мерить пустоту.
				loaded, err := s.Load()
				if err != nil {
					b.Fatal(err)
				}
				if got := len(loaded.Entries()); got != n {
					b.Fatalf("загружено %d предметов, хочу %d", got, n)
				}

				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if _, err := s.Load(); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

// BenchmarkEncode и BenchmarkDecode меряют формат отдельно от хранилищ.
func BenchmarkEncode(b *testing.B) {
	for _, n := range benchSizes {
		inv := benchInventory(b, n)
		b.Run(fmt.Sprintf("items=%d", n), func(b *testing.B) {
			var buf bytes.Buffer
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf.Reset()
				if err := encode(&buf, inv); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDecode(b *testing.B) {
	for _, n := range benchSizes {
		var buf bytes.Buffer
		if err := encode(&buf, benchInventory(b, n)); err != nil {
			b.Fatal(err)
		}
		data := buf.Bytes()

		b.Run(fmt.Sprintf("items=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := decode(bytes.NewReader(data)); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
