//go:build integration

package repository_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	dbgen "github.com/fabri-molinelli-w/opl/backend/internal/db/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/fabri-molinelli-w/opl/backend/internal/repository"
	"github.com/fabri-molinelli-w/opl/backend/internal/scheduler"
	"github.com/fabri-molinelli-w/opl/backend/internal/service"
)

// CU-03 CA-1/CA-2 + CU-10 CA-1..CA-4 contra PostgreSQL real y un servicio HTTP de prueba.
func TestMonitoringOpensAutoIncidentAndLifecycle(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, "TRUNCATE incident_events, incidents, health_checks CASCADE"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	var userID string
	if err := pool.QueryRow(ctx, `INSERT INTO users (email, name, password_hash, role)
		VALUES ('oncall@example.com', 'Guardia Demo', 'x', 'sysadmin')
		ON CONFLICT (lower(email)) DO UPDATE SET name = EXCLUDED.name RETURNING id::text`).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	var healthy atomic.Bool
	healthy.Store(true)
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if healthy.Load() {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer target.Close()

	queries := dbgen.New(pool)
	servers := service.NewServerService(repository.NewServerRepository(queries, pool), repository.NewServiceRepository(queries), discardLogger())
	catalog := service.NewServiceCatalog(repository.NewServiceRepository(queries), repository.NewServerRepository(queries, pool), discardLogger())
	srv, err := servers.Create(ctx, service.CreateServerInput{Hostname: "lab-01.lab.local", IP: "192.0.2.50", Environment: "dev"})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	svc, err := catalog.Create(ctx, service.CreateServiceInput{
		ServerID: srv.ID, Name: "lab-nginx", Type: "web", Port: 80, CheckType: "http", CheckTarget: target.URL,
	})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	monitoring := service.NewMonitoringService(repository.NewMonitoringRepository(queries, pool), scheduler.NewChecker(), 3, discardLogger())
	incidents := service.NewIncidentService(repository.NewIncidentRepository(queries, pool), discardLogger())

	check, err := monitoring.RunCheck(ctx, svc.ID)
	if err != nil || check.Status != domain.StatusUp {
		t.Fatalf("first check: %+v err=%v", check, err)
	}

	// Se "apaga" el servicio: 2 fallos no abren incidente, el 3ro sí, el 4to no duplica.
	healthy.Store(false)
	for i := 1; i <= 4; i++ {
		if _, err := monitoring.RunCheck(ctx, svc.ID); err != nil {
			t.Fatalf("down check %d: %v", i, err)
		}
		page, err := incidents.List(ctx, domain.IncidentFilter{ServiceID: svc.ID})
		if err != nil {
			t.Fatal(err)
		}
		want := int64(0)
		if i >= 3 {
			want = 1
		}
		if page.Total != want {
			t.Fatalf("after %d DOWN checks: %d incidents, want %d", i, page.Total, want)
		}
	}
	current, err := catalog.Get(ctx, svc.ID)
	if err != nil || current.CurrentStatus != domain.StatusDown {
		t.Fatalf("service status = %s, want DOWN (err %v)", current.CurrentStatus, err)
	}
	page, _ := incidents.List(ctx, domain.IncidentFilter{ServiceID: svc.ID})
	auto := page.Items[0]
	if !auto.AutoCreated || auto.Severity != "SEV2" || auto.Status != domain.IncidentOpen || auto.Title != "Servicio lab-nginx caído" {
		t.Fatalf("unexpected auto incident: %+v", auto)
	}

	// CA-2 de CU-03: vuelve a UP y se registra el evento, sin resolver solo.
	healthy.Store(true)
	if check, err := monitoring.RunCheck(ctx, svc.ID); err != nil || check.Status != domain.StatusUp {
		t.Fatalf("recovery check: %+v err=%v", check, err)
	}
	detail, err := incidents.Get(ctx, auto.ID)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Status != domain.IncidentOpen || detail.Events[len(detail.Events)-1].Body != "El servicio volvió a responder." {
		t.Fatalf("recovery not recorded: status=%s events=%+v", detail.Status, detail.Events)
	}
	checks, err := monitoring.ListChecks(ctx, svc.ID, 10)
	if err != nil || len(checks) != 6 || checks[0].Status != domain.StatusUp {
		t.Fatalf("history: %d checks err=%v", len(checks), err)
	}

	// CA-2 de CU-10: open → closed directo → 409.
	if _, err := incidents.Transition(ctx, auto.ID, domain.IncidentClosed, nil, userID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("open → closed: expected conflict, got %v", err)
	}
	if _, err := incidents.Transition(ctx, auto.ID, domain.IncidentInProgress, nil, userID); err != nil {
		t.Fatalf("open → in_progress: %v", err)
	}
	// CA-1: resolver sin causa raíz → 409.
	blank := "   "
	if _, err := incidents.Transition(ctx, auto.ID, domain.IncidentResolved, &blank, userID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("resolve without root cause: expected conflict, got %v", err)
	}
	rc := "Contenedor detenido manualmente durante la prueba"
	resolved, err := incidents.Transition(ctx, auto.ID, domain.IncidentResolved, &rc, userID)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	// CA-4: tiempo de resolución calculado.
	if resolved.ResolvedAt == nil || resolved.ResolutionMinutes() == nil || resolved.RootCause != rc {
		t.Fatalf("resolution not recorded: %+v", resolved)
	}
	closed, err := incidents.Transition(ctx, auto.ID, domain.IncidentClosed, nil, userID)
	if err != nil || closed.ClosedAt == nil {
		t.Fatalf("close: %+v err=%v", closed, err)
	}
	statusChanges := 0
	for _, e := range closed.Events {
		if e.Kind == domain.EventStatusChange {
			statusChanges++
			if e.AuthorName != "Guardia Demo" {
				t.Errorf("event author = %q", e.AuthorName)
			}
		}
	}
	if statusChanges != 3 {
		t.Errorf("expected 3 status_change events (RN-3), got %d", statusChanges)
	}

	// Con el anterior cerrado, una nueva caída abre otro incidente automático.
	healthy.Store(false)
	for i := 0; i < 3; i++ {
		if _, err := monitoring.RunCheck(ctx, svc.ID); err != nil {
			t.Fatal(err)
		}
	}
	page, _ = incidents.List(ctx, domain.IncidentFilter{ServiceID: svc.ID, OnlyActive: true})
	if page.Total != 1 {
		t.Fatalf("expected a new active auto incident after closing the previous one, got %d", page.Total)
	}

	// Alta manual con servicio inexistente → 400.
	if _, err := incidents.Create(ctx, domain.NewIncident{
		Title: "Lentitud general", Severity: "SEV3", ServiceID: "00000000-0000-4000-8000-000000000000", AuthorID: userID,
	}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("unknown service: expected validation error, got %v", err)
	}
}
