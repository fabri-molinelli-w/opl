package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	gen "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/go-chi/chi/v5"
)

type fakePinger struct {
	err error
}

func (f fakePinger) Ping(context.Context) error {
	return f.err
}

func newTestRouter(t *testing.T, p Pinger) http.Handler {
	t.Helper()
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	r := chi.NewRouter()
	gen.HandlerFromMux(gen.NewStrictHandler(NewServer(Deps{DB: p, Log: log}), nil), r)
	return r
}

func TestGetHealth(t *testing.T) {
	tests := []struct {
		name       string
		pinger     Pinger
		wantStatus int
		wantBody   gen.Health
	}{
		{
			name:       "200 ok/up when db is up",
			pinger:     fakePinger{err: nil},
			wantStatus: http.StatusOK,
			wantBody:   gen.Health{Status: gen.Ok, Database: gen.Up},
		},
		{
			name:       "503 degraded/down when db is down",
			pinger:     fakePinger{err: errors.New("connection refused")},
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   gen.Health{Status: gen.Degraded, Database: gen.Down},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRouter(t, tt.pinger)

			req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}

			var resp gen.Health
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if resp.Status != tt.wantBody.Status {
				t.Errorf("expected status %q, got %q", tt.wantBody.Status, resp.Status)
			}
			if resp.Database != tt.wantBody.Database {
				t.Errorf("expected database %q, got %q", tt.wantBody.Database, resp.Database)
			}
		})
	}
}
