package auth

import (
	"context"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

type contextKey struct{}

// WithUser stores the authenticated user in the context.
func WithUser(ctx context.Context, user domain.User) context.Context {
	return context.WithValue(ctx, contextKey{}, user)
}

// UserFromContext returns the authenticated user stored by the middleware.
func UserFromContext(ctx context.Context) (domain.User, bool) {
	user, ok := ctx.Value(contextKey{}).(domain.User)
	return user, ok
}
