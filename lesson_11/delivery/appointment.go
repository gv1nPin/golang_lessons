package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"clinic/entity"
)

type Usecase interface {
	Book(ctx context.Context, a entity.Appointment) (int, error)
	Get(ctx context.Context, id int) (entity.Appointment, error)
}

type Handler struct {
	uc Usecase
}

func New(uc Usecase) *Handler {
	return &Handler{uc: uc}
}

type bookRequest struct {
	DoctorID int       `json:"doctor_id"`
	Patient  string    `json:"patient"`
	Phone    string    `json:"phone"`
	Time     time.Time `json:"time"`
}

type bookResponse struct {
	ID int `json:"id"`
}

type appointmentResponse struct {
	ID       int       `json:"id"`
	DoctorID int       `json:"doctor_id"`
	Patient  string    `json:"patient"`
	Time     time.Time `json:"time"`
}

func (h *Handler) Book(w http.ResponseWriter, r *http.Request) {
	var req bookRequest
	body := json.NewDecoder(r.Body)
	if body.Decode(&req) != nil {
		http.Error(w, "плохой JSON",
			http.StatusBadRequest)
		return
	}
	if req.DoctorID == 0 || req.Patient == "" ||
		req.Phone == "" || req.Time.IsZero() {
		http.Error(w, "не хватает полей",
			http.StatusBadRequest)
		return
	}
	a := entity.Appointment{
		DoctorID: req.DoctorID,
		Patient:  req.Patient,
		Phone:    req.Phone,
		Time:     req.Time,
	}
	id, err := h.uc.Book(r.Context(), a)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated,
		bookResponse{ID: id})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id должен быть числом",
			http.StatusBadRequest)
		return
	}
	a, err := h.uc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, appointmentResponse{
		ID:       a.ID,
		DoctorID: a.DoctorID,
		Patient:  a.Patient,
		Time:     a.Time,
	})
}

func statusOf(err error) int {
	switch {
	case errors.Is(err, entity.ErrInPast):
		return http.StatusBadRequest
	case errors.Is(err, entity.ErrBusy):
		return http.StatusConflict
	case errors.Is(err, entity.ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func writeError(w http.ResponseWriter, err error) {
	status := statusOf(err)
	if status == http.StatusInternalServerError {
		log.Printf("internal error: %v", err)
		http.Error(w, "ошибка сервера", status)
		return
	}
	http.Error(w, err.Error(), status)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write response: %v", err)
	}
}
