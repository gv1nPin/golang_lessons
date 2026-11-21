package fathandler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type Handler struct {
	db *sql.DB
}

func New(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Book(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DoctorID int       `json:"doctor_id"`
		Patient  string    `json:"patient"`
		Time     time.Time `json:"time"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "плохой JSON", http.StatusBadRequest)
		return
	}
	if req.DoctorID == 0 || req.Time.IsZero() {
		http.Error(w, "не хватает полей", http.StatusBadRequest)
		return
	}
	if req.Time.Before(time.Now()) {
		http.Error(w, "в прошлое нельзя", http.StatusBadRequest)
		return
	}
	var n int
	err := h.db.QueryRow(
		`SELECT count(*) FROM appointments WHERE doctor_id = $1 AND at = $2`,
		req.DoctorID, req.Time).Scan(&n)
	if err != nil {
		http.Error(w, "ошибка сервера", http.StatusInternalServerError)
		return
	}
	if n > 0 {
		http.Error(w, "время занято", http.StatusConflict)
		return
	}
	var id int
	err = h.db.QueryRow(
		`INSERT INTO appointments (doctor_id, patient, at) VALUES ($1, $2, $3) RETURNING id`,
		req.DoctorID, req.Patient, req.Time).Scan(&id)
	if err != nil {
		http.Error(w, "ошибка сервера", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}
