package handlers

import (
	"context"

	gen "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) GetHealth(ctx context.Context, _ gen.GetHealthRequestObject) (gen.GetHealthResponseObject, error) {
	reqID := middleware.GetReqID(ctx)

	if err := s.db.Ping(ctx); err != nil {
		s.log.Error("database ping failed", "request_id", reqID, "error", err)
		return gen.GetHealth503JSONResponse{
			ServiceUnavailableJSONResponse: gen.ServiceUnavailableJSONResponse{
				Status:   gen.Degraded,
				Database: gen.Down,
			},
		}, nil
	}

	return gen.GetHealth200JSONResponse{
		Status:   gen.Ok,
		Database: gen.Up,
	}, nil
}
