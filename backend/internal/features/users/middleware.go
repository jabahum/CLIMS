package users

import (
	"errors"

	"clims-backend/internal/features/rbac"
	"clims-backend/internal/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

// ActiveUser resolves identity and permissions from PostgreSQL for each request.
func ActiveUser(repo *Repository) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, ok := c.Locals("user_id").(int)
		if !ok || id <= 0 {
			return response.Error(c, 401, "unauthorized", nil)
		}
		user, err := repo.GetUserByID(c.Context(), id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return response.Error(c, 401, "unauthorized", nil)
			}
			return response.Error(c, 500, "failed to load account", nil)
		}
		if user.Status != "active" {
			return response.Error(c, 403, "account is disabled", nil)
		}
		c.Locals(rbac.PrincipalKey, user)
		return c.Next()
	}
}

func CurrentUser(c fiber.Ctx) *User {
	user, _ := c.Locals(rbac.PrincipalKey).(*User)
	return user
}
