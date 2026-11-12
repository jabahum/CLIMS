package users

import (
	"clims-backend/internal/features/rbac"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(router fiber.Router, handler *Handler, jwtMiddleware, activeUser fiber.Handler) {
	users := router.Group("/users", jwtMiddleware, activeUser)
	users.Get("/", rbac.RequirePermission("users:read"), handler.ListUsers)
	users.Get("/:id", rbac.RequirePermission("users:read"), handler.GetUser)
	users.Post("/", rbac.RequirePermission("users:create"), rbac.RequirePermission("roles:assign"), handler.CreateUser)
	users.Patch("/:id/role", rbac.RequirePermission("roles:assign"), handler.AssignRole)
	users.Patch("/:id/status", rbac.RequirePermission("users:update"), handler.SetStatus)
}
