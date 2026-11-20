package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLend(t *testing.T) {
	srv := httptest.NewServer(newMux())
	defer srv.Close()
	post := func(body string) *http.Response {
		resp, err := http.Post(srv.URL+"/loans", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		return resp
	}
	steps := []struct {
		name string
		body string
		want int
	}{
		{"плохой JSON", `{`, 400},
		{"нет читателя", `{"book_id": 1}`, 400},
		{"книга 1", `{"book_id": 1, "reader_id": 10}`, 201},
		{"книга 1 уже выдана", `{"book_id": 1, "reader_id": 20}`, 409},
		{"книга 2", `{"book_id": 2, "reader_id": 10}`, 201},
		{"книга 3", `{"book_id": 3, "reader_id": 10}`, 201},
		{"четвёртая книга", `{"book_id": 4, "reader_id": 10}`, 409},
		{"другой читатель", `{"book_id": 4, "reader_id": 20}`, 201},
	}
	for _, s := range steps {
		resp := post(s.body)
		if resp.StatusCode != s.want {
			t.Fatalf("%s: code = %d, want %d", s.name, resp.StatusCode, s.want)
		}
		if s.want == 201 {
			var out map[string]string
			json.NewDecoder(resp.Body).Decode(&out)
			due, err := time.Parse("2006-01-02", out["due"])
			if err != nil || due.Before(time.Now().AddDate(0, 0, 13)) {
				t.Fatalf("%s: due = %q, want today+14", s.name, out["due"])
			}
		}
		resp.Body.Close()
	}
}
