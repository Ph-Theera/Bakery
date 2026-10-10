package line

import (
	"bakery/internal/usecases/composite"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type Sender struct {
	Client           *http.Client
	Token, Recipient string
}

func (s Sender) Send(ctx context.Context, m composite.Message) error {
	raw, err := json.Marshal(map[string]any{"to": s.Recipient, "messages": []map[string]string{{"type": "text", "text": m.Text}}})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.line.me/v2/bot/message/push", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.Token)
	req.Header.Set("X-Line-Retry-Key", m.ID)
	res, err := s.Client.Do(req)
	if err != nil {
		return errors.New("LINE request failed")
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode == 200 || res.StatusCode == 409 && res.Header.Get("X-Line-Accepted-Request-Id") != "" {
		return nil
	}
	return errors.New("LINE request rejected")
}
