//go:build integration

package repository_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/fabri-molinelli-w/opl/backend/db"
	dbgen "github.com/fabri-molinelli-w/opl/backend/internal/db/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/fabri-molinelli-w/opl/backend/internal/repository"
	"github.com/fabri-molinelli-w/opl/backend/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func integrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	t.Cleanup(pool.Close)

	sqlDB := stdlib.OpenDBFromPool(pool)
	t.Cleanup(func() { _ = sqlDB.Close() })
	goose.SetBaseFS(db.Migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}
	if err := goose.UpContext(ctx, sqlDB, "migrations"); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	if _, err := pool.Exec(ctx, "TRUNCATE services, servers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("truncate inventory: %v", err)
	}
	return pool
}

func TestInventoryFiltersPaginationAndSoftDelete(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	queries := dbgen.New(pool)
	serverRepo := repository.NewServerRepository(queries, pool)
	serviceRepo := repository.NewServiceRepository(queries)
	servers := service.NewServerService(serverRepo, serviceRepo, discardLogger())
	services := service.NewServiceCatalog(serviceRepo, serverRepo, discardLogger())

	prod, err := servers.Create(ctx, service.CreateServerInput{
		Hostname: "web-01.example.com", IP: "192.0.2.10", Environment: "prod",
		OS: "debian-12", Tags: []string{"web", "critical"},
	})
	if err != nil {
		t.Fatalf("create prod server: %v", err)
	}
	dev, err := servers.Create(ctx, service.CreateServerInput{
		Hostname: "dev-01.example.com", IP: "192.0.2.11", Environment: "dev",
		OS: "debian-12", Tags: []string{"web"},
	})
	if err != nil {
		t.Fatalf("create dev server: %v", err)
	}

	if _, err := services.Create(ctx, service.CreateServiceInput{
		ServerID: prod.ID, Name: "API", Type: "web", Port: 443,
		CheckType: "http", CheckTarget: "https://web-01.example.com/health",
	}); err != nil {
		t.Fatalf("create prod service: %v", err)
	}
	if _, err := services.Create(ctx, service.CreateServiceInput{
		ServerID: dev.ID, Name: "Worker", Type: "queue", Port: 5672,
		CheckType: "tcp", CheckTarget: "dev-01.example.com:5672",
	}); err != nil {
		t.Fatalf("create dev service: %v", err)
	}

	t.Run("filters combine", func(t *testing.T) {
		page, err := servers.List(ctx, domain.ServerFilter{Environment: "prod", Tag: "critical"})
		if err != nil {
			t.Fatalf("list servers: %v", err)
		}
		if page.Total != 1 || len(page.Items) != 1 || page.Items[0].Hostname != "web-01.example.com" {
			t.Fatalf("unexpected filtered result: %#v", page)
		}

		page, err = servers.List(ctx, domain.ServerFilter{Q: "dev-01", Environment: "prod"})
		if err != nil {
			t.Fatalf("list servers: %v", err)
		}
		if page.Total != 0 || len(page.Items) != 0 {
			t.Fatalf("expected no combined match, got %#v", page)
		}
	})

	t.Run("hostname uniqueness is case-insensitive", func(t *testing.T) {
		_, err := servers.Create(ctx, service.CreateServerInput{
			Hostname: "WEB-01.EXAMPLE.COM", IP: "192.0.2.12", Environment: "prod",
		})
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("expected ErrConflict, got %v", err)
		}
	})

	t.Run("pagination reports total", func(t *testing.T) {
		page, err := servers.List(ctx, domain.ServerFilter{Page: 2, PageSize: 1})
		if err != nil {
			t.Fatalf("list servers: %v", err)
		}
		if page.Total != 2 || page.Page != 2 || page.PageSize != 1 || len(page.Items) != 1 {
			t.Fatalf("unexpected pagination: %#v", page)
		}
	})

	t.Run("service filter by server and environment", func(t *testing.T) {
		page, err := services.List(ctx, domain.ServiceFilter{Environment: "prod"})
		if err != nil {
			t.Fatalf("list services: %v", err)
		}
		if page.Total != 1 || page.Items[0].Name != "API" {
			t.Fatalf("unexpected services: %#v", page)
		}

		page, err = services.List(ctx, domain.ServiceFilter{ServerID: dev.ID})
		if err != nil {
			t.Fatalf("list services: %v", err)
		}
		if page.Total != 1 || page.Items[0].Name != "Worker" {
			t.Fatalf("unexpected services: %#v", page)
		}
	})

	t.Run("soft deleting a server hides its services", func(t *testing.T) {
		if err := servers.Delete(ctx, prod.ID); err != nil {
			t.Fatalf("delete server: %v", err)
		}
		if _, err := servers.Get(ctx, prod.ID); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
		page, err := services.List(ctx, domain.ServiceFilter{})
		if err != nil {
			t.Fatalf("list services: %v", err)
		}
		if page.Total != 1 || page.Items[0].Name != "Worker" {
			t.Fatalf("expected only dev service, got %#v", page)
		}
	})
}

func TestDeleteServiceIsSoft(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	queries := dbgen.New(pool)
	serverRepo := repository.NewServerRepository(queries, pool)
	serviceRepo := repository.NewServiceRepository(queries)
	servers := service.NewServerService(serverRepo, serviceRepo, discardLogger())
	services := service.NewServiceCatalog(serviceRepo, serverRepo, discardLogger())

	server, err := servers.Create(ctx, service.CreateServerInput{
		Hostname: "queue-01.example.com", IP: "203.0.113.20", Environment: "staging",
	})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	item, err := services.Create(ctx, service.CreateServiceInput{
		ServerID: server.ID, Name: "Queue", Type: "queue", Port: 5672,
		CheckType: "tcp", CheckTarget: "queue-01.example.com:5672",
	})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	if err := services.Delete(ctx, item.ID); err != nil {
		t.Fatalf("delete service: %v", err)
	}
	if _, err := services.Get(ctx, item.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
