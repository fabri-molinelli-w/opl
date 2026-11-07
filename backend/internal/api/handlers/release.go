package handlers

import (
	"context"
	"time"

	gen "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

// ReleaseService is the CU-06 (certificados) and CU-07 (deploys) use cases consumed by the handlers.
type ReleaseService interface {
	ListCertificates(ctx context.Context, status string) ([]domain.Certificate, error)
	CreateCertificate(ctx context.Context, domainName string, port int) (domain.Certificate, error)
	CheckCertificate(ctx context.Context, id string) (domain.Certificate, error)
	DeleteCertificate(ctx context.Context, id string) error
	ListDeploys(ctx context.Context, f domain.DeployFilter) ([]domain.Deploy, error)
	CreateManualDeploy(ctx context.Context, d domain.NewDeploy) (string, error)
	ReceiveWebhook(ctx context.Context, token, serviceName string, d domain.NewDeploy) (string, error)
}

func toGenCertificate(c domain.Certificate) gen.Certificate {
	out := gen.Certificate{
		Id: c.ID, Domain: c.Domain, Port: c.Port, Issuer: c.Issuer, NotAfter: c.NotAfter, DaysLeft: c.DaysLeft,
		Status: gen.CertificateStatus(c.Status), LastCheckedAt: c.LastCheckedAt, CreatedAt: c.CreatedAt,
	}
	if c.Error != "" {
		e := c.Error
		out.Error = &e
	}
	return out
}

func (s *Server) ListCertificates(ctx context.Context, request gen.ListCertificatesRequestObject) (gen.ListCertificatesResponseObject, error) {
	items, err := s.releases.ListCertificates(ctx, stringValue((*string)(request.Params.Status)))
	if err != nil {
		return nil, err
	}
	out := make([]gen.Certificate, 0, len(items))
	for _, c := range items {
		out = append(out, toGenCertificate(c))
	}
	return gen.ListCertificates200JSONResponse(out), nil
}

func (s *Server) CreateCertificate(ctx context.Context, request gen.CreateCertificateRequestObject) (gen.CreateCertificateResponseObject, error) {
	if request.Body == nil {
		return nil, domain.ErrValidation
	}
	c, err := s.releases.CreateCertificate(ctx, request.Body.Domain, intValue(request.Body.Port))
	if err != nil {
		return nil, err
	}
	return gen.CreateCertificate201JSONResponse(toGenCertificate(c)), nil
}

func (s *Server) DeleteCertificate(ctx context.Context, request gen.DeleteCertificateRequestObject) (gen.DeleteCertificateResponseObject, error) {
	if err := s.releases.DeleteCertificate(ctx, request.Id); err != nil {
		return nil, err
	}
	return gen.DeleteCertificate204Response{}, nil
}

func (s *Server) CheckCertificate(ctx context.Context, request gen.CheckCertificateRequestObject) (gen.CheckCertificateResponseObject, error) {
	c, err := s.releases.CheckCertificate(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return gen.CheckCertificate200JSONResponse(toGenCertificate(c)), nil
}

func (s *Server) ListDeploys(ctx context.Context, request gen.ListDeploysRequestObject) (gen.ListDeploysResponseObject, error) {
	p := request.Params
	items, err := s.releases.ListDeploys(ctx, domain.DeployFilter{
		ServiceID: stringValue(p.ServiceId), Environment: stringValue((*string)(p.Environment)), From: p.From, To: p.To,
	})
	if err != nil {
		return nil, err
	}
	out := make([]gen.Deploy, 0, len(items))
	for _, d := range items {
		out = append(out, gen.Deploy{
			Id: d.ID, ServiceId: d.ServiceID, ServiceName: d.ServiceName, Version: d.Version,
			Environment: gen.Environment(d.Environment), Author: d.Author, Result: gen.DeployResult(d.Result),
			Notes: d.Notes, Source: gen.DeploySource(d.Source), DeployedAt: d.DeployedAt,
		})
	}
	return gen.ListDeploys200JSONResponse(out), nil
}

func (s *Server) CreateDeploy(ctx context.Context, request gen.CreateDeployRequestObject) (gen.CreateDeployResponseObject, error) {
	b := request.Body
	if b == nil {
		return nil, domain.ErrValidation
	}
	var deployedAt *time.Time
	if b.DeployedAt != nil {
		deployedAt = b.DeployedAt
	}
	id, err := s.releases.CreateManualDeploy(ctx, domain.NewDeploy{
		ServiceID: b.ServiceId, Version: b.Version, Author: stringValue(b.Author), Result: string(b.Result),
		Notes: stringValue(b.Notes), DeployedAt: deployedAt,
	})
	if err != nil {
		return nil, err
	}
	return gen.CreateDeploy201JSONResponse(gen.DeployCreated{Id: id}), nil
}

func (s *Server) ReceiveDeployWebhook(ctx context.Context, request gen.ReceiveDeployWebhookRequestObject) (gen.ReceiveDeployWebhookResponseObject, error) {
	b := request.Body
	if b == nil {
		return nil, domain.ErrValidation
	}
	id, err := s.releases.ReceiveWebhook(ctx, request.Params.XDeployToken, b.ServiceName, domain.NewDeploy{
		Version: b.Version, Environment: string(b.Environment), Author: stringValue(b.Author),
		Result: string(b.Result), Notes: stringValue(b.Notes),
	})
	if err != nil {
		return nil, err
	}
	return gen.ReceiveDeployWebhook201JSONResponse(gen.DeployCreated{Id: id}), nil
}
