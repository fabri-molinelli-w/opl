package service

import (
	"context"
	"crypto/subtle"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

// ReleaseRepository is the persistence port of CU-06 (certificados) and CU-07 (deploys).
type ReleaseRepository interface {
	CreateCertificate(ctx context.Context, domainName string, port int) (domain.Certificate, error)
	GetCertificate(ctx context.Context, id string) (domain.Certificate, error)
	ListCertificates(ctx context.Context, status string) ([]domain.Certificate, error)
	CertificatesDue(ctx context.Context, before time.Time) ([]string, error)
	SaveCertificateCheck(ctx context.Context, id string, c domain.CertificateCheck) error
	DeleteCertificate(ctx context.Context, id string) error

	ServiceEnvironment(ctx context.Context, serviceID string) (string, error)
	FindServiceForDeploy(ctx context.Context, name, environment string) (string, error)
	CreateDeploy(ctx context.Context, d domain.NewDeploy) (string, error)
	ListDeploys(ctx context.Context, f domain.DeployFilter) ([]domain.Deploy, error)
}

// CertificateReader reads the leaf certificate of a TLS endpoint.
type CertificateReader func(ctx context.Context, host string, port int) (issuer string, notAfter time.Time, err error)

// ReadTLSCertificate opens a TLS connection only to read the certificate (read-only, RN-1 de CU-06).
// InsecureSkipVerify a propósito: se quiere conocer el vencimiento aun de certificados
// autofirmados o ya vencidos; no se envía ningún dato por la conexión.
func ReadTLSCertificate(ctx context.Context, host string, port int) (string, time.Time, error) {
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: 10 * time.Second},
		Config:    &tls.Config{ServerName: host, InsecureSkipVerify: true, MinVersion: tls.VersionTLS12}, //nolint:gosec // solo lectura del vencimiento
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return "", time.Time{}, err
	}
	defer func() { _ = conn.Close() }()
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return "", time.Time{}, errors.New("conexión no TLS")
	}
	certs := tlsConn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return "", time.Time{}, errors.New("el servidor no presentó certificado")
	}
	issuer := certs[0].Issuer.CommonName
	if issuer == "" && len(certs[0].Issuer.Organization) > 0 {
		issuer = certs[0].Issuer.Organization[0]
	}
	return issuer, certs[0].NotAfter, nil
}

var domainPattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9.-]{0,251}[a-zA-Z0-9])?$`)

var validResults = map[string]bool{"success": true, "failed": true, "rollback": true}

// ReleaseService implements CU-06 and CU-07.
type ReleaseService struct {
	repo         ReleaseRepository
	read         CertificateReader
	webhookToken string
	now          func() time.Time
	log          *slog.Logger
}

func NewReleaseService(repo ReleaseRepository, read CertificateReader, webhookToken string, log *slog.Logger) *ReleaseService {
	return &ReleaseService{repo: repo, read: read, webhookToken: webhookToken, now: time.Now, log: log}
}

// --- CU-06 ---

func (s *ReleaseService) ListCertificates(ctx context.Context, status string) ([]domain.Certificate, error) {
	return s.repo.ListCertificates(ctx, status)
}

func (s *ReleaseService) CreateCertificate(ctx context.Context, domainName string, port int) (domain.Certificate, error) {
	domainName = strings.ToLower(strings.TrimSpace(domainName))
	if !domainPattern.MatchString(domainName) {
		return domain.Certificate{}, domain.NewFieldError(domain.ErrValidation, "El dominio no es válido", "domain", "formato inválido")
	}
	if port == 0 {
		port = 443
	}
	if port < 1 || port > 65535 {
		return domain.Certificate{}, domain.NewFieldError(domain.ErrValidation, "El puerto no es válido", "port", "fuera de rango")
	}
	cert, err := s.repo.CreateCertificate(ctx, domainName, port)
	if errors.Is(err, domain.ErrConflict) {
		return domain.Certificate{}, domain.NewFieldError(domain.ErrConflict, "Ese dominio y puerto ya se vigilan", "domain", "duplicado")
	}
	if err != nil {
		return domain.Certificate{}, err
	}
	s.log.Info("certificate added", "certificate_id", cert.ID, "domain", domainName, "port", port)
	return s.CheckCertificate(ctx, cert.ID)
}

// CheckCertificate reads the certificate now and stores the result. A failed read is not
// an API error: it is recorded as status unknown with the reason (RN-4).
func (s *ReleaseService) CheckCertificate(ctx context.Context, id string) (domain.Certificate, error) {
	cert, err := s.repo.GetCertificate(ctx, id)
	if err != nil {
		return domain.Certificate{}, err
	}
	result := domain.CertificateCheck{Status: domain.CertUnknown}
	issuer, notAfter, readErr := s.read(ctx, cert.Domain, cert.Port)
	if readErr != nil {
		result.Error = fmt.Sprintf("No se pudo leer el certificado: %v", readErr)
		s.log.Warn("certificate check failed", "certificate_id", id, "domain", cert.Domain, "error", readErr)
	} else {
		days := int(notAfter.Sub(s.now()).Hours() / 24)
		if notAfter.Before(s.now()) && days == 0 {
			days = -1
		}
		result.Issuer, result.NotAfter, result.DaysLeft = issuer, &notAfter, &days
		result.Status = domain.CertificateStatusFor(days)
	}
	if err := s.repo.SaveCertificateCheck(ctx, id, result); err != nil {
		return domain.Certificate{}, err
	}
	return s.repo.GetCertificate(ctx, id)
}

func (s *ReleaseService) DeleteCertificate(ctx context.Context, id string) error {
	return s.repo.DeleteCertificate(ctx, id)
}

// CheckDueCertificates re-checks every certificate not checked in the last 24 h (RN-3: chequeo diario).
func (s *ReleaseService) CheckDueCertificates(ctx context.Context, now time.Time) error {
	ids, err := s.repo.CertificatesDue(ctx, now.Add(-24*time.Hour))
	if err != nil {
		return err
	}
	for _, id := range ids {
		if _, err := s.CheckCertificate(ctx, id); err != nil {
			s.log.Error("scheduled certificate check failed", "certificate_id", id, "error", err)
		}
	}
	return nil
}

// --- CU-07 ---

func (s *ReleaseService) ListDeploys(ctx context.Context, f domain.DeployFilter) ([]domain.Deploy, error) {
	return s.repo.ListDeploys(ctx, f)
}

func validateDeploy(d *domain.NewDeploy) error {
	d.Version = strings.TrimSpace(d.Version)
	d.Author = strings.TrimSpace(d.Author)
	d.Notes = strings.TrimSpace(d.Notes)
	if d.Version == "" || len(d.Version) > 100 {
		return domain.NewFieldError(domain.ErrValidation, "La versión es obligatoria (máx. 100 caracteres)", "version", "inválida")
	}
	if !validResults[d.Result] {
		return domain.NewFieldError(domain.ErrValidation, "Resultado inválido", "result", "valor no permitido")
	}
	if len(d.Author) > 200 || len(d.Notes) > 2000 {
		return domain.NewError(domain.ErrValidation, "Autor o notas demasiado largos")
	}
	return nil
}

// CreateManualDeploy registers a deploy from the UI; the environment is the one of the service's server.
func (s *ReleaseService) CreateManualDeploy(ctx context.Context, d domain.NewDeploy) (string, error) {
	if err := validateDeploy(&d); err != nil {
		return "", err
	}
	env, err := s.repo.ServiceEnvironment(ctx, d.ServiceID)
	if errors.Is(err, domain.ErrNotFound) {
		return "", domain.NewFieldError(domain.ErrValidation, "El servicio no existe", "serviceId", "inexistente")
	}
	if err != nil {
		return "", err
	}
	d.Environment, d.Source = env, "manual"
	id, err := s.repo.CreateDeploy(ctx, d)
	if err == nil {
		s.log.Info("deploy registered", "deploy_id", id, "service_id", d.ServiceID, "source", "manual")
	}
	return id, err
}

// ReceiveWebhook registers a deploy sent by the CI/CD pipeline (RN-2: token compartido, comparación
// en tiempo constante; sin token configurado el webhook queda deshabilitado).
func (s *ReleaseService) ReceiveWebhook(ctx context.Context, token, serviceName string, d domain.NewDeploy) (string, error) {
	if s.webhookToken == "" || subtle.ConstantTimeCompare([]byte(token), []byte(s.webhookToken)) != 1 {
		s.log.Warn("deploy webhook rejected: invalid token")
		return "", domain.ErrUnauthenticated
	}
	if err := validateDeploy(&d); err != nil {
		return "", err
	}
	serviceID, err := s.repo.FindServiceForDeploy(ctx, strings.TrimSpace(serviceName), d.Environment)
	if err != nil {
		return "", err
	}
	d.ServiceID, d.Source, d.DeployedAt = serviceID, "webhook", nil
	id, err := s.repo.CreateDeploy(ctx, d)
	if err == nil {
		s.log.Info("deploy registered", "deploy_id", id, "service_id", serviceID, "source", "webhook")
	}
	return id, err
}
