package entity

import (
	"errors"
	"time"
)

var (
	ErrInPast   = errors.New("время прошло")
	ErrBusy     = errors.New("время занято")
	ErrNotFound = errors.New("не найдено")
)

type Appointment struct {
	ID       int
	DoctorID int
	Patient  string
	Phone    string
	Time     time.Time
}
