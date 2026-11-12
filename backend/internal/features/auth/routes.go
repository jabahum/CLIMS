package auth

import (
	"time"

	"clims-backend/internal/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func RegisterRoutes(router fiber.Router, handler *Handler, jwtMiddleware, activeUser fiber.Handler) {
	auth := router.Group("/auth")
	auth.Post("/register", authRateLimit(false), handler.Register)
	auth.Post("/login", authRateLimit(true), handler.Login)
	auth.Get("/me", jwtMiddleware, activeUser, handler.Me)
}

// Per-process limits; forwarded headers are not trusted as client identity.
func authRateLimit(skipSuccessful bool) fiber.Handler {
	return limiter.New(limiter.Config{
		Max: 10, Expiration: time.Minute, SkipSuccessfulRequests: skipSuccessful,
		LimitReached: func(c fiber.Ctx) error { return response.Error(c, 429, "too many attempts; try again later", nil) },
	})
}
