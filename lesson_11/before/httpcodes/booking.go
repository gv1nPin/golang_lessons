package httpcodes

import (
	"context"
	"net/http"
	"time"

	"clinic/entity"
)

type Repository interface {
	IsBusy(ctx context.Context, doctorID int, at time.Time) (bool, error)
	Save(ctx context.Context, a entity.Appointment) (int, error)
}

type Booking struct {
	repo Repository
}

func New(repo Repository) *Booking {
	return &Booking{repo: repo}
}

func (u *Booking) Book(ctx context.Context, a entity.Appointment) (int, int) {
	if a.Time.Before(time.Now()) {
		return 0, http.StatusBadRequest
	}
	busy, err := u.repo.IsBusy(ctx, a.DoctorID, a.Time)
	if err != nil {
		return 0, http.StatusInternalServerError
	}
	if busy {
		return 0, http.StatusConflict
	}
	id, err := u.repo.Save(ctx, a)
	if err != nil {
		return 0, http.StatusInternalServerError
	}
	return id, http.StatusCreated
}
