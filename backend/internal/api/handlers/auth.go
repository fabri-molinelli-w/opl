package handlers

import (
	"context"

	gen "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/auth"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

func (s *Server) Login(ctx context.Context, request gen.LoginRequestObject) (gen.LoginResponseObject, error) {
	if request.Body == nil {
		return nil, domain.NewError(domain.ErrValidation, "el cuerpo de la petición es obligatorio")
	}

	user, token, err := s.auth.Login(ctx, string(request.Body.Email), request.Body.Password)
	if err != nil {
		return nil, err
	}

	return gen.Login200JSONResponse{
		Token: token,
		User:  toGenUser(user),
	}, nil
}

func (s *Server) GetCurrentUser(ctx context.Context, _ gen.GetCurrentUserRequestObject) (gen.GetCurrentUserResponseObject, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthenticated
	}
	return gen.GetCurrentUser200JSONResponse(toGenUser(user)), nil
}

func (s *Server) ChangeOwnPassword(ctx context.Context, request gen.ChangeOwnPasswordRequestObject) (gen.ChangeOwnPasswordResponseObject, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthenticated
	}
	if request.Body == nil {
		return nil, domain.NewError(domain.ErrValidation, "el cuerpo de la petición es obligatorio")
	}

	if err := s.auth.ChangePassword(ctx, user.ID, request.Body.CurrentPassword, request.Body.NewPassword); err != nil {
		return nil, err
	}
	return gen.ChangeOwnPassword204Response{}, nil
}
