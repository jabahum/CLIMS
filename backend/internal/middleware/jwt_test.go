package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func TestJWTSubject(t *testing.T) {
	for _, tc := range []struct {
		name    string
		subject any
		want    int
	}{
		{"valid", "1465", 200},
		{"missing", nil, 401},
		{"invalid", "abc", 401},
		{"negative", "-1", 401},
		{"numeric", 1465, 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(NewJWTMiddleware("test-secret"))
			app.Get("/", func(c fiber.Ctx) error {
				if c.Locals("user_id") != 1465 {
					t.Fatal("incorrect user ID")
				}
				return c.SendStatus(200)
			})
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"sub": tc.subject, "exp": time.Now().Add(time.Hour).Unix(),
			}).SignedString([]byte("test-secret"))
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			res, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != tc.want {
				t.Fatalf("got %d, want %d", res.StatusCode, tc.want)
			}
		})
	}
}

func TestJWTRejectsInvalidTokens(t *testing.T) {
	for _, tc := range []struct {
		name    string
		expires any
		key     string
		method  jwt.SigningMethod
	}{
		{"missing expiration", nil, "test-secret", jwt.SigningMethodHS256},
		{"expired", time.Now().Add(-time.Minute).Unix(), "test-secret", jwt.SigningMethodHS256},
		{"wrong signature", time.Now().Add(time.Hour).Unix(), "other-secret", jwt.SigningMethodHS256},
		{"wrong algorithm", time.Now().Add(time.Hour).Unix(), "test-secret", jwt.SigningMethodHS384},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(NewJWTMiddleware("test-secret"))
			app.Get("/", func(c fiber.Ctx) error { return c.SendStatus(200) })
			claims := jwt.MapClaims{"sub": "1"}
			if tc.expires != nil {
				claims["exp"] = tc.expires
			}
			token, err := jwt.NewWithClaims(tc.method, claims).SignedString([]byte(tc.key))
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			res, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != 401 {
				t.Fatalf("got %d want 401", res.StatusCode)
			}
		})
	}
}
