package domain

import "time"

// Certificate status (CU-06).
const (
	CertUnknown  = "unknown"
	CertOK       = "ok"
	CertWarning  = "warning"
	CertCritical = "critical"
)

// Certificate is a TLS endpoint whose expiry OPL watches (CU-06).
type Certificate struct {
	ID            string
	Domain        string
	Port          int
	Issuer        string
	NotAfter      *time.Time
	DaysLeft      *int
	Status        string
	Error         string
	LastCheckedAt *time.Time
	CreatedAt     time.Time
}

// CertificateCheck is the outcome of reading a remote certificate.
type CertificateCheck struct {
	Issuer   string
	NotAfter *time.Time
	DaysLeft *int
	Status   string
	Error    string
}

// CertificateStatusFor applies RN-2 of CU-06: ok (> 30 días), warning (<= 30), critical (<= 7 o vencido).
func CertificateStatusFor(daysLeft int) string {
	switch {
	case daysLeft <= 7:
		return CertCritical
	case daysLeft <= 30:
		return CertWarning
	default:
		return CertOK
	}
}

// Deploy is a release of a service registered manually or by the CI/CD webhook (CU-07).
type Deploy struct {
	ID          string
	ServiceID   string
	ServiceName string
	Version     string
	Environment string
	Author      string
	Result      string
	Notes       string
	Source      string
	DeployedAt  time.Time
}

// NewDeploy is the data needed to register a deploy.
type NewDeploy struct {
	ServiceID   string
	Version     string
	Environment string
	Author      string
	Result      string
	Notes       string
	Source      string
	DeployedAt  *time.Time
}

// DeployFilter narrows the deploy history.
type DeployFilter struct {
	ServiceID   string
	Environment string
	From        *time.Time
	To          *time.Time
}
