package middleware

import (
	"log"

	jwtware "github.com/gofiber/contrib/jwt/v3"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func NewJWTMiddleware(secret string) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(secret)},
		SuccessHandler: func(c fiber.Ctx) error {
			// Optional: Extract claims and store in locals for handlers to use
			user := c.Locals("user").(*jwt.Token)
			claims := user.Claims.(jwt.MapClaims)
			c.Locals("user_id", claims["sub"])
			return c.Next()
		},
		ErrorHandler: func(c fiber.Ctx, err error) error {
			log.Printf("JWT authentication failed: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: missing or invalid token",
			})
		},
	})
}
