package middleware

import (
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
)

// NewSlogLogger initializes a JSON structured logger middleware using Go's built-in slog
func NewSlogLogger() fiber.Handler {
	// Configure JSON handler targeting stdout
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(jsonHandler)

	return func(c fiber.Ctx) error {
		start := time.Now()

		// Execute next handler in the chain
		err := c.Next()

		// Calculate request metrics
		duration := time.Since(start)
		status := c.Response().StatusCode()

		// Emit structured log entry
		logger.Info("incoming request",
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.Duration("latency", duration),
			slog.String("ip", c.IP()),
		)

		return err
	}
}
