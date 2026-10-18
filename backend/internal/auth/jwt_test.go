package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

const testSecret = "test-secret-with-at-least-32-characters"

func TestTokenIssuer_RoundTrip(t *testing.T) {
	issuer := NewTokenIssuer(testSecret, time.Hour)
	user := domain.User{ID: "user-1", Email: "user@example.com", Role: domain.RoleSoporte}

	token, err := issuer.Issue(user)
	if err != nil {
		t.Fatalf("unexpected error issuing token: %v", err)
	}

	claims, err := issuer.Parse(token)
	if err != nil {
		t.Fatalf("unexpected error parsing token: %v", err)
	}
	if claims.Subject != user.ID {
		t.Errorf("expected subject %q, got %q", user.ID, claims.Subject)
	}
	if claims.Role != user.Role {
		t.Errorf("expected role %q, got %q", user.Role, claims.Role)
	}
	if claims.Email != user.Email {
		t.Errorf("expected email %q, got %q", user.Email, claims.Email)
	}
}

func TestTokenIssuer_RejectsGarbage(t *testing.T) {
	issuer := NewTokenIssuer(testSecret, time.Hour)
	if _, err := issuer.Parse("not-a-token"); !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("expected ErrUnauthenticated, got %v", err)
	}
}

func TestTokenIssuer_RejectsExpired(t *testing.T) {
	issuer := NewTokenIssuer(testSecret, -time.Minute)
	token, err := issuer.Issue(domain.User{ID: "user-1", Role: domain.RoleAdmin})
	if err != nil {
		t.Fatalf("unexpected error issuing token: %v", err)
	}
	if _, err := issuer.Parse(token); !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("expected ErrUnauthenticated for expired token, got %v", err)
	}
}

func TestTokenIssuer_RejectsOtherSecret(t *testing.T) {
	issuer := NewTokenIssuer(testSecret, time.Hour)
	other := NewTokenIssuer("another-secret-with-at-least-32-characters", time.Hour)

	token, err := issuer.Issue(domain.User{ID: "user-1", Role: domain.RoleAdmin})
	if err != nil {
		t.Fatalf("unexpected error issuing token: %v", err)
	}
	if _, err := other.Parse(token); !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("expected ErrUnauthenticated for foreign secret, got %v", err)
	}
}
