package repository

import (
	"sync"

	"clinic/homework/fat/entity"
)

type Memory struct {
	mu          sync.Mutex
	loansByBook map[int]entity.Loan
}

func NewMemory() *Memory {
	return &Memory{
		loansByBook: make(map[int]entity.Loan),
	}
}

func (m *Memory) IsTaken(bookID int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, taken := m.loansByBook[bookID]
	return taken
}

func (m *Memory) CountByReader(readerID int) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for _, l := range m.loansByBook {
		if l.ReaderID == readerID {
			count++
		}
	}
	return count
}

func (m *Memory) Save(loan entity.Loan) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.loansByBook[loan.BookID] = loan
}
