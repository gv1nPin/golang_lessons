package storage

import (
	"bytes"
	"fmt"

	"golang_lessons/lesson_6/game/inventory"
)

// MemoryStore хранит инвентарь в памяти и подставляется вместо FileStore без изменений
// в вызывающем коде. Чтобы правила совпадали с файловым хранилищем, он хранит не сам инвентарь,
// а текст в том же формате и проходит через те же encode и decode. Поэтому:
//   - пустое хранилище даёт пустой инвентарь;
//   - Save проверяет все предметы и при ошибке (например, ErrNotStorable) оставляет прежние данные;
//   - Load каждый раз создаёт новый независимый инвентарь;
//   - повреждённые данные дают *ParseItemError с номером и содержимым строки.
type MemoryStore struct {
	data []byte
}

// NewMemoryStore создаёт пустое хранилище.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

// NewMemoryStoreFrom создаёт хранилище с готовыми данными в текстовом формате.
// Нужно, например, чтобы подставить повреждённые данные.
func NewMemoryStoreFrom(text string) *MemoryStore {
	return &MemoryStore{data: []byte(text)}
}

// Load загружает инвентарь из памяти.
func (s *MemoryStore) Load() (*inventory.Inventory, error) {
	inv, err := decode(bytes.NewReader(s.data))
	if err != nil {
		return nil, fmt.Errorf("load from memory: %w", err)
	}
	return inv, nil
}

// Save сохраняет инвентарь в память. Прежние данные заменяются только после успешного кодирования.
func (s *MemoryStore) Save(inv *inventory.Inventory) error {
	var buf bytes.Buffer
	if err := encode(&buf, inv); err != nil {
		return fmt.Errorf("save to memory: %w", err)
	}
	s.data = buf.Bytes()
	return nil
}
