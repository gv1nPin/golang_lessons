package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"clinic/delivery"
	"clinic/notifier"
	"clinic/repository"
	"clinic/usecase"
)

const (
	defaultAddr   = ":8080"
	defaultSMSURL = "http://localhost:9090/send"
)

func main() {
	addr := env("ADDR", defaultAddr)
	smsURL := env("SMS_URL", defaultSMSURL)

	repo := repository.NewMemory()
	sms := notifier.NewSMS(smsURL)
	booking := usecase.New(repo, sms)
	handler := delivery.New(booking)
	routes := handler.Routes()

	server := &http.Server{
		Addr:         addr,
		Handler:      delivery.Logging(routes),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
