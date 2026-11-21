package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SMS struct {
	url    string
	client *http.Client
}

func NewSMS(url string) *SMS {
	return &SMS{
		url:    url,
		client: &http.Client{Timeout: 2 * time.Second},
	}
}

type smsRequest struct {
	Phone string `json:"phone"`
	Text  string `json:"text"`
}

func (s *SMS) Send(ctx context.Context, phone, text string) error {
	body, err := json.Marshal(smsRequest{Phone: phone, Text: text})
	if err != nil {
		return fmt.Errorf("encode sms: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build sms request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send sms: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sms service answered %d", resp.StatusCode)
	}
	return nil
}
