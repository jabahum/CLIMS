package health

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(app *fiber.App, handler *Handler) {
	healthGroup := app.Group("/health")
	healthGroup.Get("/live", handler.Liveness)
	healthGroup.Get("/ready", handler.Readiness)
}
