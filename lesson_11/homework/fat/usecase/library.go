package usecase

import (
	"time"

	"clinic/homework/fat/entity"
)

type Repository interface {
	IsTaken(bookID int) bool
	CountByReader(readerID int) int
	Save(loan entity.Loan)
}

type Library struct {
	repo Repository
}

func New(repo Repository) *Library {
	return &Library{repo: repo}
}

func (u *Library) Lend(bookID, readerID int) (time.Time, error) {
	if u.repo.IsTaken(bookID) {
		return time.Time{}, entity.ErrTaken
	}
	if u.repo.CountByReader(readerID) >= 3 {
		return time.Time{}, entity.ErrTooMany
	}

	due := time.Now().AddDate(0, 0, 14)
	u.repo.Save(entity.Loan{
		BookID:   bookID,
		ReaderID: readerID,
		Due:      due,
	})

	return due, nil
}
