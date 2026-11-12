package rbac

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(router fiber.Router, handler *Handler, jwtMiddleware, activeUser fiber.Handler) {
	roles := router.Group("/roles", jwtMiddleware, activeUser)
	roles.Get("/", RequirePermission("roles:read"), handler.ListRoles)
	roles.Put("/:id/permissions", RequirePermission("roles:manage"), handler.SetRolePermissions)
	router.Get("/permissions", jwtMiddleware, activeUser, RequirePermission("roles:read"), handler.ListPermissions)
}
