package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"clinic/entity"
)

type Repository interface {
	IsBusy(
		ctx context.Context,
		doctorID int,
		at time.Time,
	) (bool, error)
	Save(
		ctx context.Context,
		a entity.Appointment,
	) (int, error)
	Get(
		ctx context.Context,
		id int,
	) (entity.Appointment, error)
}

type Notifier interface {
	Send(
		ctx context.Context,
		phone, text string,
	) error
}

type Booking struct {
	repo     Repository
	notifier Notifier
}

func New(repo Repository, notifier Notifier) *Booking {
	return &Booking{repo: repo, notifier: notifier}
}

func (u *Booking) Book(
	ctx context.Context,
	a entity.Appointment,
) (int, error) {
	if a.Time.Before(time.Now()) {
		return 0, entity.ErrInPast
	}
	busy, err := u.repo.IsBusy(
		ctx, a.DoctorID, a.Time)
	if err != nil {
		return 0, fmt.Errorf("is busy: %w", err)
	}
	if busy {
		return 0, entity.ErrBusy
	}
	id, err := u.repo.Save(ctx, a)
	if err != nil {
		return 0, fmt.Errorf("save: %w", err)
	}
	text := "Вы записаны на " +
		a.Time.Format("02.01 15:04")
	err = u.notifier.Send(ctx, a.Phone, text)
	if err != nil {
		log.Printf("sms not sent: %v", err)
	}
	return id, nil
}

func (u *Booking) Get(
	ctx context.Context,
	id int,
) (entity.Appointment, error) {
	return u.repo.Get(ctx, id)
}
