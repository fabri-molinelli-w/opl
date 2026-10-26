//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	dbgen "github.com/fabri-molinelli-w/opl/backend/internal/db/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/repository"
)

// CU-04 CA-1: los números del dashboard coinciden con una consulta directa a la base.
// Corre al final (zz_) para encontrar datos de los otros tests de integración.
func TestDashboardMatchesDirectQueries(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	// integrationPool vacía el inventario: se vuelven a crear datos mínimos.
	if _, err := pool.Exec(ctx, `
		INSERT INTO servers (hostname, ip, environment, location) VALUES
			('dash-a.example.com', '192.0.2.201', 'prod', 'onprem'),
			('dash-b.example.com', '192.0.2.202', 'prod', 'cloud'),
			('dash-c.example.com', '192.0.2.203', 'dev', 'onprem')`); err != nil {
		t.Fatalf("insert servers: %v", err)
	}

	summary, err := repository.NewDashboardRepository(dbgen.New(pool)).Summary(ctx, time.Now())
	if err != nil {
		t.Fatalf("summary: %v", err)
	}

	var servers, prod, cloud, docs, active, checks int
	row := pool.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM servers WHERE deleted_at IS NULL),
		(SELECT count(*) FROM servers WHERE deleted_at IS NULL AND environment = 'prod'),
		(SELECT count(*) FROM servers WHERE deleted_at IS NULL AND location = 'cloud'),
		(SELECT count(*) FROM documents WHERE deleted_at IS NULL),
		(SELECT count(*) FROM incidents WHERE status IN ('open', 'in_progress')),
		(SELECT count(*) FROM health_checks WHERE checked_at >= now() - interval '24 hours')`)
	if err := row.Scan(&servers, &prod, &cloud, &docs, &active, &checks); err != nil {
		t.Fatalf("direct counts: %v", err)
	}
	if summary.ServersTotal != servers || summary.ServersByEnvironment["prod"] != prod ||
		summary.ServersByLocation["cloud"] != cloud || summary.DocumentsTotal != docs ||
		summary.ActiveIncidents != active || summary.Checks24h != checks {
		t.Errorf("dashboard %+v does not match direct counts servers=%d prod=%d cloud=%d docs=%d active=%d checks=%d",
			summary, servers, prod, cloud, docs, active, checks)
	}
	if servers != 3 || prod != 2 || cloud != 1 {
		t.Errorf("unexpected fixture counts: servers=%d prod=%d cloud=%d", servers, prod, cloud)
	}
}
