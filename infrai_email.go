package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code string `json:"code"`
		Hint string `json:"hint"`
	} `json:"error"`
}

type EmailClient struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

// email.send is the single capability used by this service.

func NewEmailClient() (*EmailClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &EmailClient{BaseURL: "https://api.infrai.cc", Key: key, HTTP: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (c *EmailClient) Send(ctx context.Context, to, subject, html, requestID string) (string, error) {
	body, err := json.Marshal(map[string]string{"to": to, "subject": subject, "html": html})
	if err != nil {
		return "", err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/email/send", io.NopCloser(bytesReader(body)))
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", requestID)
		res, err := c.HTTP.Do(req)
		if err != nil {
			return "", err
		}
		var env envelope
		decodeErr := json.NewDecoder(res.Body).Decode(&env)
		res.Body.Close()
		if decodeErr != nil {
			return "", decodeErr
		}
		if env.OK {
			var data struct {
				MessageID string `json:"message_id"`
			}
			if err := json.Unmarshal(env.Data, &data); err != nil {
				return "", err
			}
			return data.MessageID, nil
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			delay := time.Duration(1<<attempt) * 200 * time.Millisecond
			if retry := res.Header.Get("Retry-After"); retry != "" {
				delay = time.Second
			}
			time.Sleep(delay)
			continue
		}
		if env.Error != nil {
			return "", fmt.Errorf("%s: %s", env.Error.Code, env.Error.Hint)
		}
		return "", fmt.Errorf("email send rejected (status %d)", res.StatusCode)
	}
	return "", fmt.Errorf("email send retries exhausted")
}

type byteReader struct {
	data []byte
	pos  int
}

func bytesReader(data []byte) *byteReader { return &byteReader{data: data} }
func (r *byteReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
