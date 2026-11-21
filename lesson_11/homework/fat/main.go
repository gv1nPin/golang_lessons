package main

import (
	"log"
	"net/http"

	"clinic/homework/fat/delivery"
	"clinic/homework/fat/repository"
	"clinic/homework/fat/usecase"
)

func newMux() *http.ServeMux {
	repo := repository.NewMemory()
	uc := usecase.New(repo)
	h := delivery.NewLibraryHandler(uc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /loans", h.Lend)
	return mux
}

func main() {
	log.Fatal(http.ListenAndServe(":8080", newMux()))
}
