package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"os"
)

func TestHealthHandler(t *testing.T) {
	t.Run("healthy when db is up", func(t *testing.T) {
		pool, err := pgxpool.New(t.Context(), "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
		if err != nil {
			t.Skipf("skipping: cannot connect to test db: %v", err)
		}
		defer pool.Close()

		if err := pool.Ping(t.Context()); err != nil {
			t.Skipf("skipping: cannot ping test db: %v", err)
		}

		h := NewHandler(pool, slog.New(slog.NewJSONHandler(os.Stdout, nil)))
		r := chi.NewRouter()
		r.Get("/api/health", h.ServeHTTP)

		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp Response
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Status != "healthy" {
			t.Errorf("expected status healthy, got %s", resp.Status)
		}
		if resp.Database != "up" {
			t.Errorf("expected database up, got %s", resp.Database)
		}
	})

	t.Run("unhealthy when db is down", func(t *testing.T) {
		pool, err := pgxpool.New(t.Context(), "postgres://invalid:invalid@localhost:9999/nonexistent?sslmode=disable")
		if err == nil {
			// Even if pool creation succeeded, ping should fail
			defer pool.Close()
			if err := pool.Ping(t.Context()); err == nil {
				t.Skip("skipping: unexpected successful ping to invalid db")
			}
		}

		h := NewHandler(pool, slog.New(slog.NewJSONHandler(os.Stdout, nil)))
		r := chi.NewRouter()
		r.Get("/api/health", h.ServeHTTP)

		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp Response
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Status != "unhealthy" {
			t.Errorf("expected status unhealthy, got %s", resp.Status)
		}
		if resp.Database != "down" {
			t.Errorf("expected database down, got %s", resp.Database)
		}
	})
}
