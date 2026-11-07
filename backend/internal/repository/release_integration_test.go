//go:build integration

package repository_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	dbgen "github.com/fabri-molinelli-w/opl/backend/internal/db/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/fabri-molinelli-w/opl/backend/internal/repository"
	"github.com/fabri-molinelli-w/opl/backend/internal/service"
)

// CU-06 CA-1 + CU-07 CA-1/CA-3 contra PostgreSQL real y un servidor TLS de prueba.
func TestCertificatesAndDeploys(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, "TRUNCATE certificates, deploys"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	defer tlsServer.Close()
	u, _ := url.Parse(tlsServer.URL)
	port, _ := strconv.Atoi(u.Port())

	queries := dbgen.New(pool)
	releases := service.NewReleaseService(repository.NewReleaseRepository(queries), service.ReadTLSCertificate, "token-de-prueba", discardLogger())

	cert, err := releases.CreateCertificate(ctx, "127.0.0.1", port)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	if cert.Status == domain.CertUnknown || cert.DaysLeft == nil || cert.LastCheckedAt == nil {
		t.Fatalf("certificate not checked: %+v", cert)
	}
	if _, err := releases.CreateCertificate(ctx, "127.0.0.1", port); err == nil {
		t.Fatal("duplicate domain+port must fail")
	}
	list, err := releases.ListCertificates(ctx, cert.Status)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %d err=%v", len(list), err)
	}
	if err := releases.CheckDueCertificates(ctx, time.Now().Add(48*time.Hour)); err != nil {
		t.Fatalf("due: %v", err)
	}
	if err := releases.DeleteCertificate(ctx, cert.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	servers := service.NewServerService(repository.NewServerRepository(queries, pool), repository.NewServiceRepository(queries), discardLogger())
	catalog := service.NewServiceCatalog(repository.NewServiceRepository(queries), repository.NewServerRepository(queries, pool), discardLogger())
	srv, err := servers.Create(ctx, service.CreateServerInput{Hostname: "deploy-01.lab.local", IP: "192.0.2.80", Environment: "staging", OS: "debian-12"})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	svc, err := catalog.Create(ctx, service.CreateServiceInput{ServerID: srv.ID, Name: "api-pagos", Type: "web", Port: 8080, CheckType: "tcp", CheckTarget: "192.0.2.80:8080"})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	if _, err := releases.ReceiveWebhook(ctx, "token-de-prueba", "API-PAGOS", domain.NewDeploy{Version: "v2.0.0", Environment: "staging", Result: "success", Author: "ci"}); err != nil {
		t.Fatalf("webhook: %v", err)
	}
	if _, err := releases.CreateManualDeploy(ctx, domain.NewDeploy{ServiceID: svc.ID, Version: "v2.0.1", Result: "rollback"}); err != nil {
		t.Fatalf("manual: %v", err)
	}
	from := time.Now().Add(-time.Hour)
	deploys, err := releases.ListDeploys(ctx, domain.DeployFilter{ServiceID: svc.ID, Environment: "staging", From: &from})
	if err != nil || len(deploys) != 2 || deploys[0].ServiceName != "api-pagos" {
		t.Fatalf("deploys: %+v err=%v", deploys, err)
	}
}
