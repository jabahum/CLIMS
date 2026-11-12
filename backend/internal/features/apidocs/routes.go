// Package apidocs serves the root docs/api assets without duplicating the specification.
package apidocs

import (
	"io/fs"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes accepts docs/api as its filesystem. Only public documentation
// assets are exposed; module guides and other repository files are not served.
func RegisterRoutes(router fiber.Router, assets fs.FS) {
	// Fiber treats trailing slashes as equivalent by default. Use a single
	// handler so /docs/ cannot be caught by a redirect back to itself.
	router.Get("/docs", func(c fiber.Ctx) error {
		if !strings.HasSuffix(c.Path(), "/") {
			return c.Redirect().To("/docs/")
		}
		return sendAsset(c, assets, "index.html", "text/html; charset=utf-8")
	})
	files := map[string]string{
		"/docs/openapi.json":                "openapi.json",
		"/docs/swagger-initializer.js":      "swagger-initializer.js",
		"/docs/assets/swagger-ui-bundle.js": "assets/swagger-ui-bundle.js",
		"/docs/assets/swagger-ui.css":       "assets/swagger-ui.css",
	}
	for path, name := range files {
		contentType := "application/javascript; charset=utf-8"
		switch name {
		case "index.html":
			contentType = "text/html; charset=utf-8"
		case "openapi.json":
			contentType = "application/json"
		case "assets/swagger-ui.css":
			contentType = "text/css; charset=utf-8"
		}
		router.Get(path, func(c fiber.Ctx) error {
			return sendAsset(c, assets, name, contentType)
		})
	}
}

func sendAsset(c fiber.Ctx, assets fs.FS, name, contentType string) error {
	data, err := fs.ReadFile(assets, name)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	c.Set(fiber.HeaderContentType, contentType)
	c.Set(fiber.HeaderCacheControl, "no-cache")
	return c.Send(data)
}
