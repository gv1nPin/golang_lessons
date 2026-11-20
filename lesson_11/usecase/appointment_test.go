package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"clinic/entity"
)

type fakeRepo struct {
	busy  bool
	saved []entity.Appointment
}

func (f *fakeRepo) IsBusy(
	_ context.Context, _ int, _ time.Time,
) (bool, error) {
	return f.busy, nil
}

func (f *fakeRepo) Save(
	_ context.Context, a entity.Appointment,
) (int, error) {
	f.saved = append(f.saved, a)
	return len(f.saved), nil
}

func (f *fakeRepo) Get(
	_ context.Context, id int,
) (entity.Appointment, error) {
	if id < 1 || id > len(f.saved) {
		return entity.Appointment{}, entity.ErrNotFound
	}
	return f.saved[id-1], nil
}

type fakeNotifier struct {
	err  error
	sent int
}

func (f *fakeNotifier) Send(_ context.Context, _, _ string) error {
	f.sent++
	return f.err
}

func TestBook(t *testing.T) {
	tomorrow := time.Now().Add(24 * time.Hour)
	yesterday := time.Now().Add(-24 * time.Hour)
	smsDown := errors.New("sms service is down")
	tests := []struct {
		name      string
		busy      bool
		at        time.Time
		smsErr    error
		wantErr   error
		wantSaved bool
	}{
		{"свободно", false, tomorrow, nil, nil, true},
		{"занято", true, tomorrow, nil, entity.ErrBusy, false},
		{"в прошлое", false, yesterday, nil, entity.ErrInPast, false},
		{"в прошлое и занято", true, yesterday, nil, entity.ErrInPast, false},
		{"sms не ушло", false, tomorrow, smsDown, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{busy: tt.busy}
			sms := &fakeNotifier{err: tt.smsErr}
			booking := New(repo, sms)
			a := entity.Appointment{DoctorID: 7, Patient: "Анна", Phone: "+79990000000", Time: tt.at}

			_, err := booking.Book(context.Background(), a)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if saved := len(repo.saved) == 1; saved != tt.wantSaved {
				t.Fatalf("saved = %v, want %v", saved, tt.wantSaved)
			}
			if sent := sms.sent == 1; sent != tt.wantSaved {
				t.Fatalf("sms sent = %v, want %v", sent, tt.wantSaved)
			}
		})
	}
}

func TestGet(t *testing.T) {
	repo := &fakeRepo{}
	booking := New(repo, &fakeNotifier{})
	id, err := booking.Book(context.Background(), entity.Appointment{DoctorID: 7, Time: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatalf("book: %v", err)
	}

	got, err := booking.Get(context.Background(), id)
	if err != nil || got.DoctorID != 7 {
		t.Fatalf("get %d = %+v, %v; want doctor 7", id, got, err)
	}
	if _, err := booking.Get(context.Background(), id+1); !errors.Is(err, entity.ErrNotFound) {
		t.Fatalf("get missing: err = %v, want %v", err, entity.ErrNotFound)
	}
}
