package auth

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(router fiber.Router, handler *Handler, jwtMiddleware fiber.Handler) {
	// Public Auth Group
	authGroup := router.Group("/auth")
	authGroup.Post("/register", handler.Register)
	authGroup.Post("/login", handler.Login)

	// Protected Auth Group
	protected := authGroup.Group("")
	protected.Use(jwtMiddleware)
	protected.Get("/me", handler.Me)
}
