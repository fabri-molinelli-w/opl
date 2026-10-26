package auth

import (
	"context"
	"net/http"
	"strings"

	api "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

// TokenParser validates a raw token and returns its claims.
type TokenParser interface {
	Parse(raw string) (Claims, error)
}

// UserLookup loads the current user so the middleware can enforce RN-3
// (a deactivated user is rejected even if the token is still valid).
type UserLookup interface {
	GetUserByID(ctx context.Context, id string) (domain.User, error)
}

// Authorize is the single authorization middleware. oapi-codegen (strict mode)
// passes the operationID to every StrictMiddlewareFunc, so one middleware can
// enforce the whole x-min-role policy without a role check in each handler.
func Authorize(policy Policy, tokens TokenParser, users UserLookup) api.StrictMiddlewareFunc {
	return func(next api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			if policy.IsPublic(operationID) {
				return next(ctx, w, r, request)
			}

			minRole, known := policy.MinRole(operationID)
			if !known {
				return nil, domain.ErrForbidden
			}

			user, err := authenticate(ctx, r, tokens, users)
			if err != nil {
				return nil, err
			}
			if !user.Role.AtLeast(minRole) {
				return nil, domain.ErrForbidden
			}

			ctx = WithToken(WithUser(ctx, user), bearerToken(r))
			return next(ctx, w, r, request)
		}
	}
}

const bearerPrefix = "Bearer "

func bearerToken(r *http.Request) string {
	return strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), bearerPrefix))
}

func authenticate(ctx context.Context, r *http.Request, tokens TokenParser, users UserLookup) (domain.User, error) {
	if !strings.HasPrefix(r.Header.Get("Authorization"), bearerPrefix) {
		return domain.User{}, domain.ErrUnauthenticated
	}

	claims, err := tokens.Parse(bearerToken(r))
	if err != nil {
		return domain.User{}, domain.ErrUnauthenticated
	}

	user, err := users.GetUserByID(ctx, claims.Subject)
	if err != nil {
		return domain.User{}, domain.ErrUnauthenticated
	}
	if !user.Active {
		return domain.User{}, domain.ErrUnauthenticated
	}
	return user, nil
}
