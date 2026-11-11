package auth

import (
	"time"

	"clims-backend/internal/pkg/response"
	"clims-backend/internal/pkg/validator"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	Repo      *Repository
	JWTSecret string
}

func NewHandler(repo *Repository, jwtSecret string) *Handler {
	return &Handler{Repo: repo, JWTSecret: jwtSecret}
}

func (h *Handler) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body", err.Error())
	}

	if err := validator.Validate.Struct(&req); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "validation failed", validator.FormatErrors(err))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to hash password", nil)
	}

	userID, err := h.Repo.CreateUser(c.Context(), req.Username, req.Email, string(hashedPassword))
	if err != nil {
		return response.Error(c, fiber.StatusConflict, "email or username already exists", nil)
	}

	return response.Success(c, fiber.StatusCreated, "user registered successfully", fiber.Map{
		"user_id": userID,
	})
}

func (h *Handler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body", err.Error())
	}

	if err := validator.Validate.Struct(&req); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "validation failed", validator.FormatErrors(err))
	}

	user, err := h.Repo.GetUserByEmail(c.Context(), req.Email)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "invalid credentials", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "invalid credentials", nil)
	}

	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.JWTSecret))
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to generate token", nil)
	}

	return response.Success(c, fiber.StatusOK, "login successful", fiber.Map{
		"token": tokenString,
	})
}

func (h *Handler) Me(c fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	userID := int(userIDVal.(float64))
	user, err := h.Repo.GetUserByID(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "user not found", nil)
	}

	return response.Success(c, fiber.StatusOK, "user profile retrieved successfully", UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}
