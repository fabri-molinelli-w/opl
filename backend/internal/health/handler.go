package health

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Response struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

type Handler struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func NewHandler(pool *pgxpool.Pool, log *slog.Logger) *Handler {
	return &Handler{
		pool: pool,
		log:  log,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)

	resp := Response{
		Status:   "healthy",
		Database: "up",
	}

	if err := h.pool.Ping(ctx); err != nil {
		resp.Status = "unhealthy"
		resp.Database = "down"
		h.log.Error("database ping failed", "request_id", reqID, "error", err)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.log.Error("failed to encode health response", "request_id", reqID, "error", err)
	}
}
