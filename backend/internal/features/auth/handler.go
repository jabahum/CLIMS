package auth

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"clims-backend/internal/features/users"
	"clims-backend/internal/pkg/request"
	"clims-backend/internal/pkg/response"
	"clims-backend/internal/pkg/validator"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	Repo      *users.Repository
	Service   *users.Service
	JWTSecret string
}

func NewHandler(repo *users.Repository, service *users.Service, secret string) *Handler {
	return &Handler{Repo: repo, Service: service, JWTSecret: secret}
}

func (h *Handler) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := request.DecodeJSON(c, &req); err != nil {
		return response.Error(c, 400, "invalid request body", err.Error())
	}
	req.AccountInput.Normalize()
	// Send validation errors before hashing or writing to the database.
	if err := validator.Validate.Struct(req); err != nil {
		return response.Error(c, 422, "validation failed", validator.FormatErrors(err))
	}
	if len(req.Password) > 72 {
		return response.Error(c, 422, "password must not exceed 72 bytes", nil)
	}
	return h.createUser(c, req.AccountInput, 0)
}

func (h *Handler) createUser(c fiber.Ctx, req users.AccountInput, roleID int) error {
	id, err := h.Service.Create(c.Context(), req, roleID)
	if errors.Is(err, users.ErrHashPassword) {
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

func (h *Handler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := request.DecodeJSON(c, &req); err != nil {
		return response.Error(c, 400, "invalid request body", err.Error())
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if err := validator.Validate.Struct(req); err != nil {
		return response.Error(c, 422, "validation failed", validator.FormatErrors(err))
	}
	if len(req.Password) > 72 {
		return response.Error(c, 422, "password must not exceed 72 bytes", nil)
	}
	user, err := h.Repo.GetUserByEmail(c.Context(), req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response.Error(c, 401, "invalid credentials", nil)
		}
		return response.Error(c, 500, "failed to authenticate", nil)
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return response.Error(c, 401, "invalid credentials", nil)
	}
	if user.Status != "active" {
		return response.Error(c, 403, "account is disabled", nil)
	}
	claims := jwt.MapClaims{"sub": strconv.Itoa(user.ID), "iat": time.Now().Unix(), "exp": time.Now().Add(24 * time.Hour).Unix()}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(h.JWTSecret))
	if err != nil {
		return response.Error(c, 500, "failed to generate token", nil)
	}
	return response.Success(c, 200, "login successful", fiber.Map{"token": token, "token_type": "Bearer", "expires_in": 86400, "user": user})
}

func (h *Handler) Me(c fiber.Ctx) error {
	return response.Success(c, 200, "user profile retrieved successfully", users.CurrentUser(c))
}
