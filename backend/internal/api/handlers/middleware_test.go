package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fabri-molinelli-w/opl/backend/internal/auth"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

// TestProtectedEndpoints_AuthMatrix walks every operationId declared in the
// policy and checks that it rejects requests without a token (401) and that a
// soporte token is rejected (403) exactly on the operations that require more
// than soporte.
func TestProtectedEndpoints_AuthMatrix(t *testing.T) {
	issuer := auth.NewTokenIssuer(testSecret, time.Hour)
	policy := auth.NewPolicy()

	soporte := domain.User{ID: "00000000-0000-0000-0000-0000000000aa", Email: "soporte@example.com", Role: domain.RoleSoporte, Active: true}
	token, err := issuer.Issue(soporte)
	if err != nil {
		t.Fatalf("unexpected error issuing token: %v", err)
	}

	router := newAuthorizedRouter(
		t,
		fakeAuthService{},
		fakeUserService{},
		fakeLookup{users: map[string]domain.User{soporte.ID: soporte}},
		issuer,
		policy,
	)

	endpoints := []struct {
		operation string
		method    string
		path      string
		body      string
	}{
		{auth.OperationGetCurrentUser, http.MethodGet, "/api/auth/me", ""},
		{auth.OperationChangePassword, http.MethodPost, "/api/auth/me/password", `{}`},
		{auth.OperationListUsers, http.MethodGet, "/api/users", ""},
		{auth.OperationCreateUser, http.MethodPost, "/api/users", `{}`},
		{auth.OperationUpdateUser, http.MethodPatch, "/api/users/" + soporte.ID, `{}`},
	}

	for _, ep := range endpoints {
		t.Run(ep.operation+"_without_token", func(t *testing.T) {
			req := httptest.NewRequest(ep.method, ep.path, strings.NewReader(ep.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401 without token, got %d (%s)", w.Code, w.Body.String())
			}
		})

		t.Run(ep.operation+"_with_soporte", func(t *testing.T) {
			minRole, ok := policy.MinRole(ep.operation)
			if !ok {
				t.Fatalf("operation %q has no minimum role", ep.operation)
			}

			req := httptest.NewRequest(ep.method, ep.path, strings.NewReader(ep.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if minRole == domain.RoleAdmin {
				if w.Code != http.StatusForbidden {
					t.Fatalf("expected 403 for soporte on admin operation, got %d (%s)", w.Code, w.Body.String())
				}
				return
			}
			if w.Code == http.StatusForbidden {
				t.Fatalf("soporte must not be forbidden on a soporte operation, got %d", w.Code)
			}
		})
	}
}

func TestPublicEndpoints_DoNotRequireToken(t *testing.T) {
	policy := auth.NewPolicy()
	for _, op := range []string{auth.OperationGetHealth, auth.OperationLogin} {
		if !policy.IsPublic(op) {
			t.Errorf("operation %q must be public", op)
		}
	}
}
