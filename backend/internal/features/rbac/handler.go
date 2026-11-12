package rbac

import (
	"errors"

	"clims-backend/internal/pkg/request"
	"clims-backend/internal/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Handler struct{ Repo *Repository }

func NewHandler(repo *Repository) *Handler { return &Handler{Repo: repo} }
func (h *Handler) ListRoles(c fiber.Ctx) error {
	roles, err := h.Repo.ListRoles(c.Context())
	if err != nil {
		return response.Error(c, 500, "failed to list roles", nil)
	}
	return response.Success(c, 200, "roles retrieved", roles)
}

func (h *Handler) ListPermissions(c fiber.Ctx) error {
	codes, err := h.Repo.ListPermissions(c.Context())
	if err != nil {
		return response.Error(c, 500, "failed to list permissions", nil)
	}
	return response.Success(c, 200, "permissions retrieved", codes)
}

func (h *Handler) SetRolePermissions(c fiber.Ctx) error {
	id, err := request.PositiveID(c)
	if err != nil {
		return response.Error(c, 400, err.Error(), nil)
	}
	var req SetPermissionsRequest
	if err := request.DecodeJSON(c, &req); err != nil {
		return response.Error(c, 400, "invalid request body", err.Error())
	}
	if req.Permissions == nil {
		return response.Error(c, 422, "permissions must be an array", nil)
	}
	if len(*req.Permissions) > 100 {
		return response.Error(c, 422, "too many permissions", nil)
	}
	if err := h.Repo.SetRolePermissions(c.Context(), id, *req.Permissions); err != nil {
		return accessError(c, err)
	}
	return response.Success(c, 200, "role permissions updated", nil)
}

func accessError(c fiber.Ctx, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return response.Error(c, 404, "user or role not found", nil)
	}
	if errors.Is(err, ErrAdminRole) {
		return response.Error(c, 409, err.Error(), nil)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return response.Error(c, 422, "unknown permission", nil)
	}
	return response.Error(c, 500, "failed to update or retrieve access", nil)
}
