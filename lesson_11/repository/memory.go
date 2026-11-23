package repository

import (
	"context"
	"sync"
	"time"

	"clinic/entity"
)

type Memory struct {
	mu     sync.Mutex
	items  map[int]entity.Appointment
	nextID int
}

func NewMemory() *Memory {
	return &Memory{items: map[int]entity.Appointment{}}
}

func (m *Memory) IsBusy(_ context.Context, doctorID int, at time.Time) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.items {
		if a.DoctorID == doctorID && a.Time.Equal(at) {
			return true, nil
		}
	}
	return false, nil
}

func (m *Memory) Save(_ context.Context, a entity.Appointment) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	a.ID = m.nextID
	m.items[a.ID] = a
	return a.ID, nil
}

func (m *Memory) Get(_ context.Context, id int) (entity.Appointment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.items[id]
	if !ok {
		return entity.Appointment{}, entity.ErrNotFound
	}
	return a, nil
}
