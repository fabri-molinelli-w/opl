package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/fabri-molinelli-w/opl/backend/internal/auth"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

func newTestUserService(repo *fakeUserRepo) *UserService {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewUserService(repo, log)
}

func TestCreateUser_PersistsHashNotPlaintext(t *testing.T) {
	repo := newFakeUserRepo()

	user, err := newTestUserService(repo).Create(context.Background(), CreateUserInput{
		Email:    "New@Example.com",
		Name:     "  Nueva Persona  ",
		Password: "supersecret",
		Role:     domain.RoleSoporte,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createdHash == "supersecret" {
		t.Fatal("password was stored in plaintext")
	}
	if !auth.CheckPassword(repo.createdHash, "supersecret") {
		t.Error("stored hash does not verify the original password")
	}
	if repo.createdEmail != "new@example.com" {
		t.Errorf("expected normalized email, got %q", repo.createdEmail)
	}
	if user.Name != "Nueva Persona" {
		t.Errorf("expected trimmed name, got %q", user.Name)
	}
}

func TestCreateUser_ShortPassword_Validation(t *testing.T) {
	_, err := newTestUserService(newFakeUserRepo()).Create(context.Background(), CreateUserInput{
		Email: "a@b.com", Name: "A", Password: "short", Role: domain.RoleSoporte,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestCreateUser_InvalidRole_Validation(t *testing.T) {
	_, err := newTestUserService(newFakeUserRepo()).Create(context.Background(), CreateUserInput{
		Email: "a@b.com", Name: "A", Password: "supersecret", Role: domain.Role("root"),
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestCreateUser_DuplicateEmail_Conflict(t *testing.T) {
	repo := newFakeUserRepo()
	repo.createErr = domain.ErrConflict

	_, err := newTestUserService(repo).Create(context.Background(), CreateUserInput{
		Email: "a@b.com", Name: "A", Password: "supersecret", Role: domain.RoleSoporte,
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestUpdateUser_SelfDeactivate_Conflict(t *testing.T) {
	actor := domain.User{ID: "admin-1", Role: domain.RoleAdmin}
	inactive := false

	_, err := newTestUserService(newFakeUserRepo()).Update(
		context.Background(), "admin-1", domain.UserUpdate{Active: &inactive}, actor,
	)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestUpdateUser_SelfDemote_Conflict(t *testing.T) {
	actor := domain.User{ID: "admin-1", Role: domain.RoleAdmin}
	role := domain.RoleSoporte

	_, err := newTestUserService(newFakeUserRepo()).Update(
		context.Background(), "admin-1", domain.UserUpdate{Role: &role}, actor,
	)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestUpdateUser_OtherUser_Allowed(t *testing.T) {
	repo := newFakeUserRepo()
	repo.updateResult = domain.User{ID: "user-2", Role: domain.RoleSoporte, Active: false}
	actor := domain.User{ID: "admin-1", Role: domain.RoleAdmin}
	inactive := false

	user, err := newTestUserService(repo).Update(
		context.Background(), "user-2", domain.UserUpdate{Active: &inactive}, actor,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "user-2" {
		t.Errorf("expected user-2, got %q", user.ID)
	}
}
