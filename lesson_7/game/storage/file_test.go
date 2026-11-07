package storage

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// exists сообщает, существует ли путь.
func exists(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	t.Fatal(err)
	return false
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestFileStoreLoadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "нет.dat")

	inv, err := NewFileStore(path).Load()

	if err != nil {
		t.Fatalf("файла нет: err = %v, хочу nil", err)
	}
	if n := len(inv.Entries()); n != 0 {
		t.Errorf("предметов = %d, хочу 0", n)
	}
	if exists(t, path) {
		t.Error("Load не должен создавать файл")
	}
}

func TestFileStoreLoadErrors(t *testing.T) {
	t.Run("путь указывает на папку", func(t *testing.T) {
		inv, err := NewFileStore(t.TempDir()).Load()

		if err == nil {
			t.Fatal("ошибки нет")
		}
		if errors.Is(err, fs.ErrNotExist) {
			t.Errorf("папка существует, но err = %v", err)
		}
		if inv != nil {
			t.Error("при ошибке инвентарь не должен возвращаться")
		}
	})

	t.Run("ошибка содержит путь и сохраняет ParseItemError", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "inventory.dat")
		if err := os.WriteFile(path, []byte("бред\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		_, err := NewFileStore(path).Load()

		var parseErr *ParseItemError
		if !errors.As(err, &parseErr) || parseErr.Line != 1 {
			t.Fatalf("err = %v, хочу *ParseItemError для строки 1", err)
		}
		if !strings.Contains(err.Error(), path) {
			t.Errorf("в тексте ошибки нет пути %q: %q", path, err)
		}
	})
}

func TestFileStoreSave(t *testing.T) {
	t.Run("файл содержит ровно сохранённые данные и нет временного файла", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "inventory.dat")

		if err := NewFileStore(path).Save(sampleInventory(t)); err != nil {
			t.Fatal(err)
		}

		if got := readFile(t, path); got != sampleDump {
			t.Errorf("содержимое файла %q, хочу %q", got, sampleDump)
		}
		if exists(t, path+".tmp") {
			t.Error("временный файл остался")
		}
	})

	t.Run("остаток временного файла от прошлого сбоя не мешает", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "inventory.dat")
		stale := strings.Repeat("мусор от прошлого сбоя\n", 20)
		if err := os.WriteFile(path+".tmp", []byte(stale), 0o644); err != nil {
			t.Fatal(err)
		}

		if err := NewFileStore(path).Save(sampleInventory(t)); err != nil {
			t.Fatal(err)
		}

		if got := readFile(t, path); got != sampleDump {
			t.Errorf("содержимое файла %q, хочу %q", got, sampleDump)
		}
		if exists(t, path+".tmp") {
			t.Error("временный файл остался")
		}
	})

	t.Run("папки для файла не существует", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "нет-такой-папки", "inventory.dat")

		err := NewFileStore(path).Save(sampleInventory(t))

		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("err = %v, хочу fs.ErrNotExist", err)
		}
		if err != nil && !strings.Contains(err.Error(), path) {
			t.Errorf("в тексте ошибки нет пути %q: %q", path, err)
		}
		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if len(entries) != 0 {
			t.Errorf("Save создал лишние файлы: %v", entries)
		}
	})

	t.Run("путь указывает на папку", func(t *testing.T) {
		dir := t.TempDir()

		err := NewFileStore(dir).Save(sampleInventory(t))

		if err == nil {
			t.Fatal("ошибки нет")
		}
		if info, statErr := os.Stat(dir); statErr != nil || !info.IsDir() {
			t.Errorf("папка пропала или стала файлом: %v", statErr)
		}
		if exists(t, dir+".tmp") {
			t.Error("временный файл остался")
		}
	})
}

// Бонус: после ошибки посередине записи прежний файл не повреждён.
func TestWriteFileAtomic(t *testing.T) {
	boom := errors.New("диск отвалился")
	big := strings.Repeat("x", 3*4096) // больше буфера bufio, чтобы данные реально попали во временный файл

	writeString := func(s string) func(io.Writer) error {
		return func(w io.Writer) error {
			_, err := io.WriteString(w, s)
			return err
		}
	}
	writeThenFail := func(s string) func(io.Writer) error {
		return func(w io.Writer) error {
			_, _ = io.WriteString(w, s)
			return boom
		}
	}

	tests := []struct {
		name        string
		old         string // прежнее содержимое файла
		hasOld      bool
		write       func(io.Writer) error
		wantErr     error
		wantContent string // содержимое файла после вызова (если файл должен существовать)
	}{
		{"успех: новый файл", "", false, writeString("новое\n"), nil, "новое\n"},
		{"успех: поверх старого", "старое длинное содержимое\n", true, writeString("новое\n"), nil, "новое\n"},
		{"успех: запись больше буфера", "", false, writeString(big), nil, big},
		{"успех: пустая запись заменяет файл", "старое\n", true, writeString(""), nil, ""},
		{"ошибка до записи", "старое\n", true, func(io.Writer) error { return boom }, boom, "старое\n"},
		{"ошибка после небольшой записи", "старое\n", true, writeThenFail("половина"), boom, "старое\n"},
		{"ошибка после записи больше буфера", "старое\n", true, writeThenFail(big), boom, "старое\n"},
		{"ошибка, когда прежнего файла не было", "", false, writeThenFail("половина"), boom, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "inventory.dat")
			if tt.hasOld {
				if err := os.WriteFile(path, []byte(tt.old), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			err := writeFileAtomic(path, tt.write)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, хочу %v", err, tt.wantErr)
			}
			if exists(t, path+".tmp") {
				t.Error("временный файл остался")
			}
			if tt.wantErr != nil && !tt.hasOld {
				if exists(t, path) {
					t.Error("при ошибке не должен появляться новый файл")
				}
				return
			}
			if got := readFile(t, path); got != tt.wantContent {
				t.Errorf("содержимое файла %q, хочу %q", got, tt.wantContent)
			}
		})
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

	t.Run("нет прав на чтение", func(t *testing.T) {
		if err := os.Chmod(path, 0o000); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(path, 0o644) })

		inv, err := s.Load()

		if !errors.Is(err, fs.ErrPermission) {
			t.Errorf("err = %v, хочу fs.ErrPermission", err)
		}
		if inv != nil {
			t.Error("при ошибке инвентарь не должен возвращаться")
		}
	})

	t.Run("нет прав на запись в папку", func(t *testing.T) {
		if err := os.Chmod(dir, 0o500); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

		err := s.Save(sampleInventory(t))

		if !errors.Is(err, fs.ErrPermission) {
			t.Errorf("err = %v, хочу fs.ErrPermission", err)
		}
		if got := readFile(t, path); got != sampleDump {
			t.Errorf("файл изменился после неудачной записи: %q", got)
		}
	})
}
