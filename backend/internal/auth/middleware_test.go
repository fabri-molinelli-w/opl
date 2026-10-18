package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

type fakeUserLookup struct {
	users map[string]domain.User
}

func (f fakeUserLookup) GetUserByID(_ context.Context, id string) (domain.User, error) {
	user, ok := f.users[id]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func requestWithToken(token string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func TestAuthenticate_ValidToken(t *testing.T) {
	issuer := NewTokenIssuer(testSecret, time.Hour)
	user := domain.User{ID: "user-1", Email: "user@example.com", Role: domain.RoleSoporte, Active: true}
	token, err := issuer.Issue(user)
	if err != nil {
		t.Fatalf("unexpected error issuing token: %v", err)
	}
	lookup := fakeUserLookup{users: map[string]domain.User{user.ID: user}}

	got, err := authenticate(context.Background(), requestWithToken(token), issuer, lookup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("expected user %q, got %q", user.ID, got.ID)
	}
}

func TestAuthenticate_MissingHeader(t *testing.T) {
	issuer := NewTokenIssuer(testSecret, time.Hour)
	_, err := authenticate(context.Background(), requestWithToken(""), issuer, fakeUserLookup{})
	if !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("expected ErrUnauthenticated, got %v", err)
	}
}

func TestAuthenticate_DeactivatedUserRejected(t *testing.T) {
	issuer := NewTokenIssuer(testSecret, time.Hour)
	user := domain.User{ID: "user-1", Role: domain.RoleSoporte, Active: false}
	token, err := issuer.Issue(user)
	if err != nil {
		t.Fatalf("unexpected error issuing token: %v", err)
	}
	lookup := fakeUserLookup{users: map[string]domain.User{user.ID: user}}

	_, err = authenticate(context.Background(), requestWithToken(token), issuer, lookup)
	if !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("expected ErrUnauthenticated for a deactivated user, got %v", err)
	}
}
