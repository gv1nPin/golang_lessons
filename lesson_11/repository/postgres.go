package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"clinic/entity"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *Postgres {
	return &Postgres{db: db}
}

func (p *Postgres) IsBusy(ctx context.Context, doctorID int, at time.Time) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM appointments WHERE doctor_id = $1 AND at = $2)`
	var busy bool
	err := p.db.QueryRowContext(ctx, query, doctorID, at).Scan(&busy)
	if err != nil {
		return false, fmt.Errorf("select busy slot: %w", err)
	}
	return busy, nil
}

func (p *Postgres) Save(ctx context.Context, a entity.Appointment) (int, error) {
	const query = `INSERT INTO appointments (doctor_id, patient, phone, at) VALUES ($1, $2, $3, $4) RETURNING id`
	var id int
	err := p.db.QueryRowContext(ctx, query, a.DoctorID, a.Patient, a.Phone, a.Time).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert appointment: %w", err)
	}
	return id, nil
}

func (p *Postgres) Get(ctx context.Context, id int) (entity.Appointment, error) {
	const query = `SELECT id, doctor_id, patient, phone, at FROM appointments WHERE id = $1`
	var a entity.Appointment
	err := p.db.QueryRowContext(ctx, query, id).Scan(&a.ID, &a.DoctorID, &a.Patient, &a.Phone, &a.Time)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Appointment{}, entity.ErrNotFound
	}
	if err != nil {
		return entity.Appointment{}, fmt.Errorf("select appointment %d: %w", id, err)
	}
	return a, nil
}
