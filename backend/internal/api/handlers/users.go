package handlers

import (
	"context"

	gen "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/auth"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/fabri-molinelli-w/opl/backend/internal/service"
)

func (s *Server) ListUsers(ctx context.Context, _ gen.ListUsersRequestObject) (gen.ListUsersResponseObject, error) {
	users, err := s.users.List(ctx)
	if err != nil {
		return nil, err
	}

	response := make(gen.ListUsers200JSONResponse, 0, len(users))
	for _, user := range users {
		response = append(response, toGenUser(user))
	}
	return response, nil
}

func (s *Server) CreateUser(ctx context.Context, request gen.CreateUserRequestObject) (gen.CreateUserResponseObject, error) {
	if request.Body == nil {
		return nil, domain.NewError(domain.ErrValidation, "el cuerpo de la petición es obligatorio")
	}

	user, err := s.users.Create(ctx, service.CreateUserInput{
		Email:    string(request.Body.Email),
		Name:     request.Body.Name,
		Password: request.Body.Password,
		Role:     domain.Role(request.Body.Role),
	})
	if err != nil {
		return nil, err
	}
	return gen.CreateUser201JSONResponse(toGenUser(user)), nil
}

func (s *Server) UpdateUser(ctx context.Context, request gen.UpdateUserRequestObject) (gen.UpdateUserResponseObject, error) {
	actor, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthenticated
	}
	if request.Body == nil {
		return nil, domain.NewError(domain.ErrValidation, "el cuerpo de la petición es obligatorio")
	}

	var update domain.UserUpdate
	update.Name = request.Body.Name
	update.Active = request.Body.Active
	if request.Body.Email != nil {
		email := string(*request.Body.Email)
		update.Email = &email
	}
	if request.Body.Role != nil {
		role := domain.Role(*request.Body.Role)
		update.Role = &role
	}

	user, err := s.users.Update(ctx, request.Id, update, actor)
	if err != nil {
		return nil, err
	}
	return gen.UpdateUser200JSONResponse(toGenUser(user)), nil
}
