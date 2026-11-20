package delivery

import "net/http"

func (h *Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /appointments", h.Book)
	mux.HandleFunc("GET /appointments/{id}", h.Get)
	mux.HandleFunc("GET /ping", ping)
	return mux
}

func ping(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("pong"))
}
