package aiclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

func TestChatForwardsIdentityAndInternalToken(t *testing.T) {
	ai := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/chat" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-Internal-Token") != "internal" {
			t.Errorf("missing internal token")
		}
		if r.Header.Get("Authorization") != "Bearer user-jwt" || r.Header.Get("X-User-Id") != "user-1" {
			t.Errorf("user identity not forwarded: %v", r.Header)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["sessionId"] != "s1" || body["message"] != "hola" {
			t.Errorf("unexpected body %v (%v)", body, err)
		}
		_, _ = w.Write([]byte(`{"answer":"ok","sources":[{"type":"document","id":"d1","title":"Doc"}]}`))
	}))
	defer ai.Close()

	answer, err := New(ai.URL, "internal").Chat(context.Background(), domain.ChatQuestion{
		SessionID: "s1", Message: "hola", UserID: "user-1", UserToken: "user-jwt",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if answer.Answer != "ok" || len(answer.Sources) != 1 || answer.Sources[0].ID != "d1" {
		t.Errorf("unexpected answer: %+v", answer)
	}
}

func TestChatNon200IsError(t *testing.T) {
	ai := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ai.Close()

	if _, err := New(ai.URL, "internal").Chat(context.Background(), domain.ChatQuestion{}); err == nil {
		t.Fatal("expected error on 502")
	}
}
