package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

type fakeReleaseRepo struct {
	cert     domain.Certificate
	saved    domain.CertificateCheck
	deploys  []domain.NewDeploy
	services map[string]string // name|env -> id
}

func (f *fakeReleaseRepo) CreateCertificate(_ context.Context, d string, p int) (domain.Certificate, error) {
	f.cert = domain.Certificate{ID: "c1", Domain: d, Port: p, Status: domain.CertUnknown}
	return f.cert, nil
}
func (f *fakeReleaseRepo) GetCertificate(context.Context, string) (domain.Certificate, error) {
	c := f.cert
	c.Status, c.DaysLeft, c.Error = f.saved.Status, f.saved.DaysLeft, f.saved.Error
	return c, nil
}
func (f *fakeReleaseRepo) ListCertificates(context.Context, string) ([]domain.Certificate, error) {
	return nil, nil
}
func (f *fakeReleaseRepo) CertificatesDue(context.Context, time.Time) ([]string, error) {
	return nil, nil
}
func (f *fakeReleaseRepo) SaveCertificateCheck(_ context.Context, _ string, c domain.CertificateCheck) error {
	f.saved = c
	return nil
}
func (f *fakeReleaseRepo) DeleteCertificate(context.Context, string) error { return nil }
func (f *fakeReleaseRepo) ServiceEnvironment(_ context.Context, id string) (string, error) {
	if id != "s1" {
		return "", domain.ErrNotFound
	}
	return "prod", nil
}
func (f *fakeReleaseRepo) FindServiceForDeploy(_ context.Context, name, env string) (string, error) {
	if id, ok := f.services[name+"|"+env]; ok {
		return id, nil
	}
	return "", domain.ErrNotFound
}
func (f *fakeReleaseRepo) CreateDeploy(_ context.Context, d domain.NewDeploy) (string, error) {
	f.deploys = append(f.deploys, d)
	return "d1", nil
}
func (f *fakeReleaseRepo) ListDeploys(context.Context, domain.DeployFilter) ([]domain.Deploy, error) {
	return nil, nil
}

func TestCertificateStatusRule(t *testing.T) {
	cases := map[int]string{90: domain.CertOK, 31: domain.CertOK, 30: domain.CertWarning, 8: domain.CertWarning, 7: domain.CertCritical, -3: domain.CertCritical}
	for days, want := range cases {
		if got := domain.CertificateStatusFor(days); got != want {
			t.Errorf("%d días: %s, want %s", days, got, want)
		}
	}
}

func TestCreateCertificateChecksImmediately(t *testing.T) {
	now := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	repo := &fakeReleaseRepo{}
	read := func(context.Context, string, int) (string, time.Time, error) {
		return "Lab CA", now.Add(5 * 24 * time.Hour), nil
	}
	svc := NewReleaseService(repo, read, "", discardLogger())
	svc.now = func() time.Time { return now }
	cert, err := svc.CreateCertificate(context.Background(), " Lab-TLS.lab.local ", 0)
	if err != nil {
		t.Fatal(err)
	}
	if cert.Domain != "lab-tls.lab.local" || cert.Port != 443 || cert.Status != domain.CertCritical || *cert.DaysLeft != 5 {
		t.Fatalf("unexpected certificate: %+v", cert)
	}
	if _, err := svc.CreateCertificate(context.Background(), "no valido/x", 443); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("invalid domain: %v", err)
	}

	// Un dominio inalcanzable queda en unknown con el motivo (no es error de la API).
	svc.read = func(context.Context, string, int) (string, time.Time, error) {
		return "", time.Time{}, errors.New("connection refused")
	}
	cert, err = svc.CheckCertificate(context.Background(), "c1")
	if err != nil || cert.Status != domain.CertUnknown || cert.Error == "" {
		t.Fatalf("unreachable: %+v err=%v", cert, err)
	}
}

func TestDeployWebhookToken(t *testing.T) {
	repo := &fakeReleaseRepo{services: map[string]string{"lab-nginx|dev": "s1"}}
	svc := NewReleaseService(repo, nil, "secreto-de-prueba", discardLogger())
	in := domain.NewDeploy{Version: "v1.2.0", Environment: "dev", Result: "success"}
	ctx := context.Background()

	if _, err := svc.ReceiveWebhook(ctx, "otro", "lab-nginx", in); !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("bad token: %v", err)
	}
	if _, err := svc.ReceiveWebhook(ctx, "secreto-de-prueba", "inexistente", in); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unknown service: %v", err)
	}
	if _, err := svc.ReceiveWebhook(ctx, "secreto-de-prueba", "lab-nginx", in); err != nil {
		t.Fatal(err)
	}
	if got := repo.deploys[0]; got.ServiceID != "s1" || got.Source != "webhook" {
		t.Fatalf("unexpected deploy: %+v", got)
	}
	disabled := NewReleaseService(repo, nil, "", discardLogger())
	if _, err := disabled.ReceiveWebhook(ctx, "", "lab-nginx", in); !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("webhook without configured token must be rejected: %v", err)
	}
}

func TestManualDeployValidation(t *testing.T) {
	repo := &fakeReleaseRepo{}
	svc := NewReleaseService(repo, nil, "", discardLogger())
	ctx := context.Background()
	if _, err := svc.CreateManualDeploy(ctx, domain.NewDeploy{ServiceID: "s1", Version: " ", Result: "success"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("empty version: %v", err)
	}
	if _, err := svc.CreateManualDeploy(ctx, domain.NewDeploy{ServiceID: "zz", Version: "v1", Result: "success"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("unknown service: %v", err)
	}
	if _, err := svc.CreateManualDeploy(ctx, domain.NewDeploy{ServiceID: "s1", Version: "v1", Result: "rollback"}); err != nil {
		t.Fatal(err)
	}
	if got := repo.deploys[0]; got.Environment != "prod" || got.Source != "manual" {
		t.Fatalf("unexpected deploy: %+v", got)
	}
}
