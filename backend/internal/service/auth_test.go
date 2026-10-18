package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/fabri-molinelli-w/opl/backend/internal/auth"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

const testSecret = "test-secret-with-at-least-32-characters"

func newTestAuthService(repo *fakeUserRepo) *AuthService {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewAuthService(repo, auth.NewTokenIssuer(testSecret, time.Hour), log)
}

func mustHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}
	return hash
}

func TestLogin_Success(t *testing.T) {
	repo := newFakeUserRepo()
	repo.byEmail["user@example.com"] = domain.User{
		ID: "user-1", Email: "user@example.com", Role: domain.RoleSoporte, Active: true,
		PasswordHash: mustHash(t, "supersecret"),
	}

	user, token, err := newTestAuthService(repo).Login(context.Background(), "user@example.com", "supersecret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected a non-empty token")
	}
	if user.ID != "user-1" {
		t.Errorf("expected user-1, got %q", user.ID)
	}
}

func TestLogin_EmailCaseInsensitive(t *testing.T) {
	repo := newFakeUserRepo()
	repo.byEmail["user@example.com"] = domain.User{
		ID: "user-1", Email: "user@example.com", Role: domain.RoleSoporte, Active: true,
		PasswordHash: mustHash(t, "supersecret"),
	}

	if _, _, err := newTestAuthService(repo).Login(context.Background(), "  USER@Example.com ", "supersecret"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newFakeUserRepo()
	repo.byEmail["user@example.com"] = domain.User{
		ID: "user-1", Role: domain.RoleSoporte, Active: true,
		PasswordHash: mustHash(t, "supersecret"),
	}

	_, _, err := newTestAuthService(repo).Login(context.Background(), "user@example.com", "wrong-password")
	if !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("expected ErrUnauthenticated, got %v", err)
	}
}

func TestLogin_UnknownEmail(t *testing.T) {
	_, _, err := newTestAuthService(newFakeUserRepo()).Login(context.Background(), "nobody@example.com", "whatever")
	if !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("expected ErrUnauthenticated, got %v", err)
	}
}

func TestLogin_InactiveUser(t *testing.T) {
	repo := newFakeUserRepo()
	repo.byEmail["user@example.com"] = domain.User{
		ID: "user-1", Role: domain.RoleSoporte, Active: false,
		PasswordHash: mustHash(t, "supersecret"),
	}

	_, _, err := newTestAuthService(repo).Login(context.Background(), "user@example.com", "supersecret")
	if !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("expected ErrUnauthenticated for inactive user, got %v", err)
	}
}

func TestChangePassword_WrongCurrent(t *testing.T) {
	repo := newFakeUserRepo()
	repo.byID["user-1"] = domain.User{ID: "user-1", PasswordHash: mustHash(t, "supersecret")}

	err := newTestAuthService(repo).ChangePassword(context.Background(), "user-1", "wrong", "newpassword")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestChangePassword_ShortNewPassword(t *testing.T) {
	err := newTestAuthService(newFakeUserRepo()).ChangePassword(context.Background(), "user-1", "whatever", "short")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestChangePassword_Success(t *testing.T) {
	repo := newFakeUserRepo()
	repo.byID["user-1"] = domain.User{ID: "user-1", PasswordHash: mustHash(t, "supersecret")}

	if err := newTestAuthService(repo).ChangePassword(context.Background(), "user-1", "supersecret", "newpassword"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.passwordHash == "" {
		t.Fatal("expected the password hash to be updated")
	}
	if !auth.CheckPassword(repo.passwordHash, "newpassword") {
		t.Error("stored hash does not match the new password")
	}
}

func TestEnsureAdminBootstrap_CreatesWhenNone(t *testing.T) {
	repo := newFakeUserRepo()
	repo.adminCount = 0

	err := newTestAuthService(repo).EnsureAdminBootstrap(context.Background(), "admin@opl.local", "adminpassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createdEmail != "admin@opl.local" {
		t.Errorf("expected admin email to be created, got %q", repo.createdEmail)
	}
	if repo.createdRole != domain.RoleAdmin {
		t.Errorf("expected role admin, got %q", repo.createdRole)
	}
	if repo.createdHash == "adminpassword" {
		t.Error("admin password was stored in plaintext")
	}
}

func TestEnsureAdminBootstrap_SkipsWhenAdminExists(t *testing.T) {
	repo := newFakeUserRepo()
	repo.adminCount = 1

	if err := newTestAuthService(repo).EnsureAdminBootstrap(context.Background(), "admin@opl.local", "adminpassword"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createdEmail != "" {
		t.Error("bootstrap must not create another admin")
	}
}

func TestEnsureAdminBootstrap_MissingCredentials(t *testing.T) {
	repo := newFakeUserRepo()
	repo.adminCount = 0

	err := newTestAuthService(repo).EnsureAdminBootstrap(context.Background(), "", "")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}
