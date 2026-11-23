package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"clinic/entity"
)

func TestMemory(t *testing.T) {
	ctx := context.Background()
	at := time.Date(2030, 1, 2, 10, 0, 0, 0, time.UTC)
	repo := NewMemory()

	id, err := repo.Save(ctx, entity.Appointment{DoctorID: 7, Patient: "Анна", Time: at})
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	if busy, _ := repo.IsBusy(ctx, 7, at); !busy {
		t.Fatal("doctor 7 must be busy after save")
	}
	if busy, _ := repo.IsBusy(ctx, 8, at); busy {
		t.Fatal("doctor 8 must be free")
	}
	if got, err := repo.Get(ctx, id); err != nil || got.Patient != "Анна" {
		t.Fatalf("get %d = %+v, %v", id, got, err)
	}
	if _, err := repo.Get(ctx, id+1); !errors.Is(err, entity.ErrNotFound) {
		t.Fatalf("get missing: err = %v, want %v", err, entity.ErrNotFound)
	}
}
