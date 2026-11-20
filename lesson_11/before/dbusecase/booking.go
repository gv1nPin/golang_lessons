package dbusecase

import (
	"context"
	"database/sql"
	"time"

	"clinic/entity"
	"clinic/repository"
)

const dsn = "postgres://localhost/clinic"

type Booking struct {
	repo *repository.Postgres
}

func New() *Booking {
	db, _ := sql.Open("postgres", dsn)
	return &Booking{
		repo: repository.NewPostgres(db),
	}
}

func (u *Booking) Book(ctx context.Context, a entity.Appointment) (int, error) {
	if a.Time.Before(time.Now()) {
		return 0, entity.ErrInPast
	}
	busy, err := u.repo.IsBusy(ctx, a.DoctorID, a.Time)
	if err != nil {
		return 0, err
	}
	if busy {
		return 0, entity.ErrBusy
	}
	return u.repo.Save(ctx, a)
}
