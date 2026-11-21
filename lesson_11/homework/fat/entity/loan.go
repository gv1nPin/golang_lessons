package entity

import (
	"errors"
	"time"
)

var (
	ErrTaken   = errors.New("book is already taken")
	ErrTooMany = errors.New("reader already has 3 books")
)

type Loan struct {
	BookID   int
	ReaderID int
	Due      time.Time
}
