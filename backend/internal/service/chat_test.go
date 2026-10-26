package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

type fakeOpsbot struct {
	got domain.ChatQuestion
	err error
}

func (f *fakeOpsbot) Chat(_ context.Context, q domain.ChatQuestion) (domain.ChatAnswer, error) {
	f.got = q
	return domain.ChatAnswer{Answer: "ok"}, f.err
}

func chatLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func validQuestion() domain.ChatQuestion {
	return domain.ChatQuestion{SessionID: " s1 ", Message: "  ¿Qué está caído?  ", UserID: "u1", UserToken: "jwt"}
}

func TestChatAskTrimsAndForwards(t *testing.T) {
	bot := &fakeOpsbot{}
	if _, err := NewChatService(bot, chatLogger()).Ask(context.Background(), validQuestion()); err != nil {
		t.Fatal(err)
	}
	if bot.got.Message != "¿Qué está caído?" || bot.got.SessionID != "s1" {
		t.Errorf("not trimmed: %+v", bot.got)
	}
}

func TestChatAskValidation(t *testing.T) {
	tests := []struct {
		name string
		edit func(*domain.ChatQuestion)
		want error
	}{
		{"mensaje vacío", func(q *domain.ChatQuestion) { q.Message = "   " }, domain.ErrValidation},
		{"mensaje largo", func(q *domain.ChatQuestion) { q.Message = strings.Repeat("a", 4001) }, domain.ErrValidation},
		{"sin sesión", func(q *domain.ChatQuestion) { q.SessionID = "" }, domain.ErrValidation},
		{"sin token", func(q *domain.ChatQuestion) { q.UserToken = "" }, domain.ErrUnauthenticated},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := validQuestion()
			tt.edit(&q)
			_, err := NewChatService(&fakeOpsbot{}, chatLogger()).Ask(context.Background(), q)
			if !errors.Is(err, tt.want) {
				t.Errorf("expected %v, got %v", tt.want, err)
			}
		})
	}
}

func TestChatAskUnavailable(t *testing.T) {
	_, err := NewChatService(nil, chatLogger()).Ask(context.Background(), validQuestion())
	if !errors.Is(err, domain.ErrUnavailable) {
		t.Errorf("not configured: expected ErrUnavailable, got %v", err)
	}
	_, err = NewChatService(&fakeOpsbot{err: errors.New("dial tcp 10.0.0.5: refused")}, chatLogger()).
		Ask(context.Background(), validQuestion())
	if !errors.Is(err, domain.ErrUnavailable) || strings.Contains(err.Error(), "10.0.0.5") {
		t.Errorf("ai failure must map to ErrUnavailable without internal details, got %v", err)
	}
}
