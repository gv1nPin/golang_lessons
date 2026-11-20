package notifier

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSend(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{"сервис ответил 200", http.StatusOK, false},
		{"сервис ответил 500", http.StatusInternalServerError, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer server.Close()

			err := NewSMS(server.URL).Send(context.Background(), "+79990000000", "Вы записаны")

			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
