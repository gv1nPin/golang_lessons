package delivery

import (
	"encoding/json"
	"errors"
	"net/http"

	"clinic/homework/fat/entity"
	"clinic/homework/fat/usecase"
)

type LibraryHandler struct {
	uc *usecase.Library
}

func NewLibraryHandler(uc *usecase.Library) *LibraryHandler {
	return &LibraryHandler{uc: uc}
}

func (h *LibraryHandler) Lend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BookID   int `json:"book_id"`
		ReaderID int `json:"reader_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.BookID == 0 || req.ReaderID == 0 {
		http.Error(w, "плохой запрос", http.StatusBadRequest)
		return
	}

	due, err := h.uc.Lend(req.BookID, req.ReaderID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrTaken):
			http.Error(w, "книга уже выдана", http.StatusConflict)
		case errors.Is(err, entity.ErrTooMany):
			http.Error(w, "на руках уже 3 книги", http.StatusConflict)
		default:
			http.Error(w, "внутренняя ошибка", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"due": due.Format("2006-01-02")})
}
