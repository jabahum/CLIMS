package auth

import (
	"net/http/httptest"
	"strings"
	"testing"

	"clims-backend/internal/features/users"
	"clims-backend/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

func TestLoginAttemptLimit(t *testing.T) {
	app := fiber.New()
	RegisterRoutes(app.Group("/api/v1"), NewHandler(nil, nil, "test-secret"), middleware.NewJWTMiddleware("test-secret"), users.ActiveUser(nil))
	for attempt := 1; attempt <= 11; attempt++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		want := 422
		if attempt == 11 {
			want = 429
		}
		if res.StatusCode != want {
			t.Fatalf("attempt %d: got %d want %d", attempt, res.StatusCode, want)
		}
	}
}
