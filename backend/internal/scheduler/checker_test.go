package scheduler

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

// CA-4 de CU-03: clasificador UP / DEGRADED / DOWN contra un httptest.Server.
func TestCheckHTTPClassification(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
		case "/slow":
			time.Sleep(60 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		case "/redirect":
			http.Redirect(w, r, "/ok", http.StatusFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	tests := []struct {
		name      string
		path      string
		threshold int
		want      string
		wantCode  int
	}{
		{"responde lo esperado", "/ok", 1000, domain.StatusUp, 200},
		{"lento pero correcto", "/slow", 20, domain.StatusDegraded, 200},
		{"código distinto", "/error", 1000, domain.StatusDown, 500},
		{"no sigue redirecciones", "/redirect", 1000, domain.StatusDown, 302},
	}
	checker := NewChecker()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checker.Check(context.Background(), domain.CheckConfig{
				CheckType: "http", CheckTarget: srv.URL + tt.path, ExpectedStatus: 200, LatencyThresholdMs: tt.threshold,
			})
			if got.Status != tt.want {
				t.Fatalf("status = %s, want %s (error %q)", got.Status, tt.want, got.Error)
			}
			if got.HTTPStatus == nil || *got.HTTPStatus != tt.wantCode {
				t.Errorf("http status = %v, want %d", got.HTTPStatus, tt.wantCode)
			}
			if got.ResponseMs == nil {
				t.Error("response time must be recorded")
			}
		})
	}
}

func TestCheckHTTPConnectionRefusedIsDown(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()

	got := NewChecker().Check(context.Background(), domain.CheckConfig{
		CheckType: "http", CheckTarget: url, ExpectedStatus: 200, LatencyThresholdMs: 1000,
	})
	if got.Status != domain.StatusDown || got.Error == "" {
		t.Fatalf("expected DOWN with error, got %+v", got)
	}
}

func TestCheckTCP(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()
	addr := ln.Addr().String()

	checker := NewChecker()
	up := checker.Check(context.Background(), domain.CheckConfig{CheckType: "tcp", CheckTarget: addr, LatencyThresholdMs: 1000})
	if up.Status != domain.StatusUp {
		t.Fatalf("expected UP, got %+v", up)
	}
	_ = ln.Close()
	down := checker.Check(context.Background(), domain.CheckConfig{CheckType: "tcp", CheckTarget: addr, LatencyThresholdMs: 1000})
	if down.Status != domain.StatusDown {
		t.Fatalf("expected DOWN after closing the listener, got %+v", down)
	}
}

func TestClassify(t *testing.T) {
	if got := Classify(true, 10, 100, nil, ""); got.Status != domain.StatusUp {
		t.Errorf("fast ok: %s", got.Status)
	}
	if got := Classify(true, 101, 100, nil, ""); got.Status != domain.StatusDegraded {
		t.Errorf("slow ok: %s", got.Status)
	}
	if got := Classify(false, 10, 100, nil, "boom"); got.Status != domain.StatusDown || got.Error != "boom" {
		t.Errorf("failure: %+v", got)
	}
}
