package service

import (
	"context"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

type fakeUserRepo struct {
	byEmail map[string]domain.User
	byID    map[string]domain.User
	list    []domain.User

	createdEmail string
	createdHash  string
	createdRole  domain.Role
	createErr    error

	updateResult domain.User
	updateErr    error
	lastUpdate   domain.UserUpdate

	passwordHash string
	updatePwErr  error

	adminCount int64
	countErr   error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		byEmail: map[string]domain.User{},
		byID:    map[string]domain.User{},
	}
}

func (f *fakeUserRepo) GetUserByEmail(_ context.Context, email string) (domain.User, error) {
	user, ok := f.byEmail[email]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func (f *fakeUserRepo) GetUserByID(_ context.Context, id string) (domain.User, error) {
	user, ok := f.byID[id]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func (f *fakeUserRepo) ListUsers(_ context.Context) ([]domain.User, error) {
	return f.list, nil
}

func (f *fakeUserRepo) CreateUser(_ context.Context, email, name, passwordHash string, role domain.Role) (domain.User, error) {
	if f.createErr != nil {
		return domain.User{}, f.createErr
	}
	f.createdEmail = email
	f.createdHash = passwordHash
	f.createdRole = role
	return domain.User{ID: "new-id", Email: email, Name: name, PasswordHash: passwordHash, Role: role, Active: true}, nil
}

func (f *fakeUserRepo) UpdateUser(_ context.Context, _ string, update domain.UserUpdate) (domain.User, error) {
	f.lastUpdate = update
	if f.updateErr != nil {
		return domain.User{}, f.updateErr
	}
	return f.updateResult, nil
}

func (f *fakeUserRepo) UpdatePassword(_ context.Context, _ string, passwordHash string) error {
	if f.updatePwErr != nil {
		return f.updatePwErr
	}
	f.passwordHash = passwordHash
	return nil
}

func (f *fakeUserRepo) CountAdmins(_ context.Context) (int64, error) {
	return f.adminCount, f.countErr
}
