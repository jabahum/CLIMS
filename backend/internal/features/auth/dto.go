package auth

import "clims-backend/internal/features/users"

type RegisterRequest struct{ users.AccountInput }

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,max=72"`
}
