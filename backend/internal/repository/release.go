package repository

import (
	"context"
	"time"

	dbgen "github.com/fabri-molinelli-w/opl/backend/internal/db/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

// ReleaseRepository stores watched certificates (CU-06) and deploys (CU-07).
type ReleaseRepository struct {
	q *dbgen.Queries
}

func NewReleaseRepository(q *dbgen.Queries) *ReleaseRepository {
	return &ReleaseRepository{q: q}
}

func toDomainCertificate(row dbgen.GetCertificateRow) domain.Certificate {
	c := domain.Certificate{
		ID: uuidToString(row.ID), Domain: row.Domain, Port: int(row.Port), Issuer: row.Issuer,
		NotAfter: timePtr(row.NotAfter), DaysLeft: int4Ptr(row.DaysLeft), Status: row.Status,
		LastCheckedAt: timePtr(row.LastCheckedAt), CreatedAt: row.CreatedAt.Time,
	}
	if row.Error.Valid {
		c.Error = row.Error.String
	}
	return c
}

func (r *ReleaseRepository) CreateCertificate(ctx context.Context, domainName string, port int) (domain.Certificate, error) {
	row, err := r.q.CreateCertificate(ctx, dbgen.CreateCertificateParams{Domain: domainName, Port: int32(port)}) //nolint:gosec // puerto validado (1-65535)
	if err != nil {
		return domain.Certificate{}, mapError(err)
	}
	return toDomainCertificate(dbgen.GetCertificateRow(row)), nil
}

func (r *ReleaseRepository) GetCertificate(ctx context.Context, id string) (domain.Certificate, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return domain.Certificate{}, domain.ErrNotFound
	}
	row, err := r.q.GetCertificate(ctx, uid)
	if err != nil {
		return domain.Certificate{}, mapError(err)
	}
	return toDomainCertificate(row), nil
}

func (r *ReleaseRepository) ListCertificates(ctx context.Context, status string) ([]domain.Certificate, error) {
	rows, err := r.q.ListCertificates(ctx, textParam(status))
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]domain.Certificate, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainCertificate(dbgen.GetCertificateRow(row)))
	}
	return out, nil
}

func (r *ReleaseRepository) CertificatesDue(ctx context.Context, before time.Time) ([]string, error) {
	rows, err := r.q.ListCertificatesDue(ctx, ts(before))
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]string, 0, len(rows))
	for _, id := range rows {
		out = append(out, uuidToString(id))
	}
	return out, nil
}

func (r *ReleaseRepository) SaveCertificateCheck(ctx context.Context, id string, c domain.CertificateCheck) error {
	uid, err := uuidFromString(id)
	if err != nil {
		return domain.ErrNotFound
	}
	params := dbgen.UpdateCertificateCheckParams{ID: uid, Issuer: c.Issuer, Status: c.Status, DaysLeft: optionalInt4(c.DaysLeft)}
	if c.NotAfter != nil {
		params.NotAfter = ts(*c.NotAfter)
	}
	if c.Error != "" {
		params.Error = textValue(c.Error)
	}
	return mapError(r.q.UpdateCertificateCheck(ctx, params))
}

func (r *ReleaseRepository) DeleteCertificate(ctx context.Context, id string) error {
	uid, err := uuidFromString(id)
	if err != nil {
		return domain.ErrNotFound
	}
	n, err := r.q.DeleteCertificate(ctx, uid)
	if err != nil {
		return mapError(err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ReleaseRepository) ServiceEnvironment(ctx context.Context, serviceID string) (string, error) {
	uid, err := uuidFromString(serviceID)
	if err != nil {
		return "", domain.ErrNotFound
	}
	env, err := r.q.ServiceEnvironment(ctx, uid)
	if err != nil {
		return "", mapError(err)
	}
	return env, nil
}

func (r *ReleaseRepository) FindServiceForDeploy(ctx context.Context, name, environment string) (string, error) {
	id, err := r.q.FindServiceForDeploy(ctx, dbgen.FindServiceForDeployParams{Name: name, Environment: environment})
	if err != nil {
		return "", mapError(err)
	}
	return uuidToString(id), nil
}

func (r *ReleaseRepository) CreateDeploy(ctx context.Context, d domain.NewDeploy) (string, error) {
	uid, err := uuidFromString(d.ServiceID)
	if err != nil {
		return "", domain.ErrNotFound
	}
	params := dbgen.CreateDeployParams{
		ServiceID: uid, Version: d.Version, Environment: d.Environment, Author: d.Author,
		Result: d.Result, Notes: d.Notes, Source: d.Source,
	}
	if d.DeployedAt != nil {
		params.DeployedAt = ts(*d.DeployedAt)
	}
	id, err := r.q.CreateDeploy(ctx, params)
	if err != nil {
		return "", mapError(err)
	}
	return uuidToString(id), nil
}

func (r *ReleaseRepository) ListDeploys(ctx context.Context, f domain.DeployFilter) ([]domain.Deploy, error) {
	params := dbgen.ListDeploysParams{Environment: textParam(f.Environment)}
	if f.ServiceID != "" {
		uid, err := uuidFromString(f.ServiceID)
		if err != nil {
			return nil, domain.NewFieldError(domain.ErrValidation, "serviceId inválido", "serviceId", "uuid inválido")
		}
		params.ServiceID = uid
	}
	if f.From != nil {
		params.FromTs = ts(*f.From)
	}
	if f.To != nil {
		params.ToTs = ts(*f.To)
	}
	rows, err := r.q.ListDeploys(ctx, params)
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]domain.Deploy, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.Deploy{
			ID: uuidToString(row.ID), ServiceID: uuidToString(row.ServiceID), ServiceName: row.ServiceName,
			Version: row.Version, Environment: row.Environment, Author: row.Author, Result: row.Result,
			Notes: row.Notes, Source: row.Source, DeployedAt: row.DeployedAt.Time,
		})
	}
	return out, nil
}
