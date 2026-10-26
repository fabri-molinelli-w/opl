// Package aiclient talks to the internal ai-service (OPSBot). The backend is the
// only component allowed to reach it (AGENTS.md §4).
package aiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

// Client calls POST /chat on the ai-service with the internal token and the
// user's own JWT, so OPSBot tools inherit the user's permissions (ADR-004).
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func New(baseURL, internalToken string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   internalToken,
		// The LLM may chain several tool calls; keep a generous but bounded timeout.
		http: &http.Client{Timeout: 90 * time.Second},
	}
}

type chatRequest struct {
	SessionID string `json:"sessionId"`
	Message   string `json:"message"`
}

type chatResponse struct {
	Answer  string `json:"answer"`
	Sources []struct {
		Type  string `json:"type"`
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"sources"`
}

func (c *Client) Chat(ctx context.Context, q domain.ChatQuestion) (domain.ChatAnswer, error) {
	body, err := json.Marshal(chatRequest{SessionID: q.SessionID, Message: q.Message})
	if err != nil {
		return domain.ChatAnswer{}, fmt.Errorf("encode chat request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat", bytes.NewReader(body))
	if err != nil {
		return domain.ChatAnswer{}, fmt.Errorf("build chat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", c.token)
	req.Header.Set("Authorization", "Bearer "+q.UserToken)
	req.Header.Set("X-User-Id", q.UserID)

	resp, err := c.http.Do(req)
	if err != nil {
		return domain.ChatAnswer{}, fmt.Errorf("call ai-service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return domain.ChatAnswer{}, fmt.Errorf("ai-service responded %d", resp.StatusCode)
	}

	var decoded chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return domain.ChatAnswer{}, fmt.Errorf("decode chat response: %w", err)
	}
	answer := domain.ChatAnswer{Answer: decoded.Answer, Sources: make([]domain.ChatSource, 0, len(decoded.Sources))}
	for _, s := range decoded.Sources {
		answer.Sources = append(answer.Sources, domain.ChatSource{Type: s.Type, ID: s.ID, Title: s.Title})
	}
	return answer, nil
}
