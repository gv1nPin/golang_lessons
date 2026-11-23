package delivery

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"clinic/entity"
)

type fakeUsecase struct {
	err error
}

func (f *fakeUsecase) Book(_ context.Context, _ entity.Appointment) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	return 1, nil
}

func (f *fakeUsecase) Get(_ context.Context, id int) (entity.Appointment, error) {
	if f.err != nil {
		return entity.Appointment{}, f.err
	}
	return entity.Appointment{ID: id, DoctorID: 7, Patient: "Анна"}, nil
}

func TestBook(t *testing.T) {
	valid := `{"doctor_id": 7, "patient": "Анна", "phone": "+79990000000", "time": "` +
		time.Now().Add(time.Hour).Format(time.RFC3339) + `"}`
	tests := []struct {
		name   string
		body   string
		ucErr  error
		status int
	}{
		{"записались", valid, nil, http.StatusCreated},
		{"плохой JSON", `{`, nil, http.StatusBadRequest},
		{"не хватает полей", `{"doctor_id": 7}`, nil, http.StatusBadRequest},
		{"в прошлое", valid, entity.ErrInPast, http.StatusBadRequest},
		{"занято", valid, entity.ErrBusy, http.StatusConflict},
		{"база упала", valid, errors.New("connection refused"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			routes := New(&fakeUsecase{err: tt.ucErr}).Routes()
			req := httptest.NewRequest(http.MethodPost, "/appointments", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			routes.ServeHTTP(rec, req)

			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.status, rec.Body)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		ucErr  error
		status int
	}{
		{"нашли", "/appointments/1", nil, http.StatusOK},
		{"не число", "/appointments/abc", nil, http.StatusBadRequest},
		{"нет такой", "/appointments/999", entity.ErrNotFound, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			routes := New(&fakeUsecase{err: tt.ucErr}).Routes()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			routes.ServeHTTP(rec, req)

			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.status, rec.Body)
			}
		})
	}
}
