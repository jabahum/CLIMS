package users

import (
	"context"
	"errors"
	"strconv"

	"clims-backend/internal/pkg/request"
	"clims-backend/internal/pkg/response"
	"clims-backend/internal/pkg/validator"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type RoleCatalog interface {
	RoleExists(context.Context, int) (bool, error)
}
type Handler struct {
	Repo    *Repository
	Service *Service
	Roles   RoleCatalog
}

func NewHandler(repo *Repository, service *Service, roles RoleCatalog) *Handler {
	return &Handler{Repo: repo, Service: service, Roles: roles}
}

func (h *Handler) createUser(c fiber.Ctx, req AccountInput, roleID int) error {
	id, err := h.Service.Create(c.Context(), req, roleID)
	if errors.Is(err, ErrHashPassword) {
		return response.Error(c, 500, "failed to hash password", nil)
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return response.Error(c, 409, "email already exists", nil)
		}
		return response.Error(c, 500, "failed to create user", nil)
	}
	return response.Success(c, 201, "user registered successfully", fiber.Map{"user_id": id})
}

func (h *Handler) ListUsers(c fiber.Ctx) error {
	limit, offset := 50, 0
	var err error
	if raw := c.Query("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 100 {
			return response.Error(c, 400, "limit must be between 1 and 100", nil)
		}
	}
	if raw := c.Query("offset"); raw != "" {
		offset, err = strconv.Atoi(raw)
		if err != nil || offset < 0 {
			return response.Error(c, 400, "offset must be nonnegative", nil)
		}
	}
	users, err := h.Repo.ListUsers(c.Context(), limit, offset)
	if err != nil {
		return response.Error(c, 500, "failed to list users", nil)
	}
	return response.Success(c, 200, "users retrieved", fiber.Map{"users": users, "limit": limit, "offset": offset})
}

func (h *Handler) GetUser(c fiber.Ctx) error {
	id, err := request.PositiveID(c)
	if err != nil {
		return response.Error(c, 400, err.Error(), nil)
	}
	user, err := h.Repo.GetUserByID(c.Context(), id)
	if err != nil {
		return accessError(c, err)
	}
	return response.Success(c, 200, "user retrieved", user)
}

func (h *Handler) CreateUser(c fiber.Ctx) error {
	var req CreateUserRequest
	if err := request.DecodeJSON(c, &req); err != nil {
		return response.Error(c, 400, "invalid request body", err.Error())
	}
	req.AccountInput.Normalize()
	if err := validator.Validate.Struct(req); err != nil {
		return response.Error(c, 422, "validation failed", validator.FormatErrors(err))
	}
	if len(req.Password) > 72 {
		return response.Error(c, 422, "password must not exceed 72 bytes", nil)
	}
	exists, err := h.Roles.RoleExists(c.Context(), req.RoleID)
	if err != nil {
		return response.Error(c, 500, "failed to load role", nil)
	}
	if !exists {
		return response.Error(c, 422, "unknown role_id", nil)
	}
	return h.createUser(c, req.AccountInput, req.RoleID)
}

func (h *Handler) AssignRole(c fiber.Ctx) error {
	id, err := request.PositiveID(c)
	if err != nil {
		return response.Error(c, 400, err.Error(), nil)
	}
	var req AssignRoleRequest
	if err := request.DecodeJSON(c, &req); err != nil {
		return response.Error(c, 400, "invalid request body", err.Error())
	}
	if err := validator.Validate.Struct(req); err != nil {
		return response.Error(c, 422, "validation failed", validator.FormatErrors(err))
	}
	if err := h.Repo.UpdateUserAccess(c.Context(), id, &req.RoleID, nil); err != nil {
		return accessError(c, err)
	}
	return response.Success(c, 200, "role assigned", nil)
}

func (h *Handler) SetStatus(c fiber.Ctx) error {
	id, err := request.PositiveID(c)
	if err != nil {
		return response.Error(c, 400, err.Error(), nil)
	}
	var req UpdateStatusRequest
	if err := request.DecodeJSON(c, &req); err != nil {
		return response.Error(c, 400, "invalid request body", err.Error())
	}
	if err := validator.Validate.Struct(req); err != nil {
		return response.Error(c, 422, "validation failed", validator.FormatErrors(err))
	}
	if err := h.Repo.UpdateUserAccess(c.Context(), id, nil, &req.Status); err != nil {
		return accessError(c, err)
	}
	return response.Success(c, 200, "status updated", nil)
}

func accessError(c fiber.Ctx, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return response.Error(c, 404, "user or role not found", nil)
	}
	if errors.Is(err, ErrLastAdmin) {
		return response.Error(c, 409, err.Error(), nil)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return response.Error(c, 422, "unknown permission", nil)
	}
	return response.Error(c, 500, "failed to update or retrieve access", nil)
}
