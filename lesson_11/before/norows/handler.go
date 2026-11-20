package norows

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"clinic/entity"
)

type Usecase interface {
	Get(ctx context.Context, id int) (entity.Appointment, error)
}

type Handler struct {
	uc Usecase
}

func New(uc Usecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id должен быть числом", http.StatusBadRequest)
		return
	}
	a, err := h.uc.Get(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "не найдено", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "ошибка сервера", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a)
}
