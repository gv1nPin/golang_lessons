package storage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"golang_lessons/lesson_6/game/inventory"
)

// FileStore хранит инвентарь в файле. Отвечает только за работу с файлом:
// формат записи живёт в encode и decode.
type FileStore struct {
	path string
}

// NewFileStore создаёт хранилище, которое читает и пишет файл по пути path.
func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

// Load загружает инвентарь из файла. Если файла ещё нет (первый запуск), это не ошибка:
// возвращается пустой инвентарь, как и у MemoryStore, в который ещё ничего не сохраняли.
// Остальные ошибки (нет прав, повреждённая строка) возвращаются с обёрткой,
// и errors.Is / errors.As видят причину.
func (s *FileStore) Load() (*inventory.Inventory, error) {
	file, err := os.Open(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return inventory.New(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	// При чтении ошибка Close не влияет на сохранность данных, поэтому её можно проигнорировать.
	defer func() { _ = file.Close() }()

	inv, err := decode(file)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", s.path, err)
	}
	return inv, nil
}

// Save сохраняет инвентарь в файл через временный файл и замену.
//
// Данные сначала пишутся в path+".tmp", и только после успешной записи, Flush, Sync
// и Close временный файл переименовывается поверх настоящего. Поэтому ошибка посреди записи
// не портит прежний файл, а новое содержимое короче старого не оставляет «хвоста» от него.
func (s *FileStore) Save(inv *inventory.Inventory) error {
	err := writeFileAtomic(s.path, func(w io.Writer) error {
		return encode(w, inv)
	})
	if err != nil {
		return fmt.Errorf("save %s: %w", s.path, err)
	}
	return nil
}

// writeFileAtomic записывает в файл то, что пишет функция write, по схеме «временный файл + замена».
// Если на любом шаге возникла ошибка, временный файл удаляется, а path остаётся нетронутым.
func writeFileAtomic(path string, write func(io.Writer) error) (err error) {
	tmpPath := path + ".tmp"
	tmp, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	// Удаление идёт после явных Close ниже: на Windows открытый файл удалить нельзя.
	defer func() {
		if err != nil {
			_ = os.Remove(tmpPath) // уборка: ошибка удаления не важнее исходной ошибки
		}
	}()

	buffered := bufio.NewWriter(tmp)
	if err = write(buffered); err != nil {
		_ = tmp.Close() // основная ошибка уже есть; закрываем, чтобы не держать файл
		return fmt.Errorf("write: %w", err)
	}
	if err = buffered.Flush(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("flush: %w", err)
	}
	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("sync: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	if err = os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}
