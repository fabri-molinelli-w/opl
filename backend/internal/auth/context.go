package auth

import (
	"context"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

type contextKey struct{}

type tokenKey struct{}

// WithUser stores the authenticated user in the context.
func WithUser(ctx context.Context, user domain.User) context.Context {
	return context.WithValue(ctx, contextKey{}, user)
}

// UserFromContext returns the authenticated user stored by the middleware.
func UserFromContext(ctx context.Context) (domain.User, bool) {
	user, ok := ctx.Value(contextKey{}).(domain.User)
	return user, ok
}

// WithToken stores the raw bearer token of the request. It is only used to
// forward the user's identity to internal services (OPSBot), never logged.
func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey{}, token)
}

// TokenFromContext returns the raw bearer token stored by the middleware.
func TokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenKey{}).(string)
	return token, ok && token != ""
}
