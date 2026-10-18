package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gen "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/auth"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

const testSecret = "test-secret-with-at-least-32-characters"

func TestLoginHandler(t *testing.T) {
	user := domain.User{ID: "user-1", Email: "user@example.com", Name: "User", Role: domain.RoleSoporte, Active: true}

	tests := []struct {
		name       string
		service    fakeAuthService
		wantStatus int
	}{
		{
			name:       "200 on valid credentials",
			service:    fakeAuthService{loginUser: user, loginToken: "signed-token"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "401 on invalid credentials",
			service:    fakeAuthService{loginErr: domain.ErrUnauthenticated},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newAuthorizedRouter(t, tt.service, fakeUserService{}, fakeLookup{}, auth.NewTokenIssuer(testSecret, time.Hour), auth.NewPolicy())

			body := strings.NewReader(`{"email":"user@example.com","password":"supersecret"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d (%s)", tt.wantStatus, w.Code, w.Body.String())
			}

			if tt.wantStatus == http.StatusOK {
				var resp gen.LoginResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.Token != "signed-token" {
					t.Errorf("expected token to be returned, got %q", resp.Token)
				}
				if resp.User.Email != "user@example.com" {
					t.Errorf("unexpected user in response: %+v", resp.User)
				}
			}
		})
	}
}

func TestCreateUserHandler_Conflict(t *testing.T) {
	issuer := auth.NewTokenIssuer(testSecret, time.Hour)
	admin := domain.User{ID: "admin-1", Role: domain.RoleAdmin, Active: true}
	token, err := issuer.Issue(admin)
	if err != nil {
		t.Fatalf("unexpected error issuing token: %v", err)
	}

	router := newAuthorizedRouter(
		t,
		fakeAuthService{},
		fakeUserService{createErr: domain.ErrConflict},
		fakeLookup{users: map[string]domain.User{admin.ID: admin}},
		issuer,
		auth.NewPolicy(),
	)

	body := strings.NewReader(`{"email":"dup@example.com","name":"Dup","password":"supersecret","role":"soporte"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/users", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d (%s)", w.Code, w.Body.String())
	}
}
