package health

import (
	"context"
	"time"

	"clims-backend/internal/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB         *pgxpool.Pool
	Version    string
	CommitHash string
	BuildTime  string
}

func NewHandler(db *pgxpool.Pool, version, commit, buildTime string) *Handler {
	return &Handler{
		DB:         db,
		Version:    version,
		CommitHash: commit,
		BuildTime:  buildTime,
	}
}

func (h *Handler) Liveness(c fiber.Ctx) error {
	return response.Success(c, fiber.StatusOK, "service is live", fiber.Map{
		"status":      "UP",
		"version":     h.Version,
		"commit_hash": h.CommitHash,
		"build_time":  h.BuildTime,
	})
}

func (h *Handler) Readiness(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	if err := h.DB.Ping(ctx); err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "database is unreachable", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "service and database are ready", fiber.Map{
		"database": "connected",
		"status":   "UP",
		"version":  h.Version,
	})
}
