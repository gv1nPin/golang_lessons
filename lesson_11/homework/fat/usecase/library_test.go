package usecase

import (
	"errors"
	"testing"
	"time"

	"clinic/homework/fat/entity"
)

type fakeRepo struct {
	taken   bool
	onHands int
	saved   int
}

func (f *fakeRepo) IsTaken(_ int) bool {
	return f.taken
}

func (f *fakeRepo) CountByReader(_ int) int {
	return f.onHands
}

func (f *fakeRepo) Save(_ entity.Loan) {
	f.saved++
}

func TestLend(t *testing.T) {
	tests := []struct {
		name      string
		taken     bool
		onHands   int
		wantErr   error
		wantSaved int
	}{
		{
			name:      "успешная выдача",
			taken:     false,
			onHands:   0,
			wantErr:   nil,
			wantSaved: 1,
		},
		{
			name:      "книга уже выдана",
			taken:     true,
			onHands:   0,
			wantErr:   entity.ErrTaken,
			wantSaved: 0,
		},
		{
			name:      "на руках уже 3 книги",
			taken:     false,
			onHands:   3,
			wantErr:   entity.ErrTooMany,
			wantSaved: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{
				taken:   tt.taken,
				onHands: tt.onHands,
			}
			lib := New(repo)

			before := time.Now().AddDate(0, 0, 14)
			due, err := lib.Lend(1, 10)
			after := time.Now().AddDate(0, 0, 14)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if repo.saved != tt.wantSaved {
				t.Fatalf("saved = %d, want %d", repo.saved, tt.wantSaved)
			}
			if tt.wantErr == nil {
				if due.Before(before) || due.After(after) {
					t.Fatalf("due = %v, want between %v and %v", due, before, after)
				}
			}
		})
	}
}
