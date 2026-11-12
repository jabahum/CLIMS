package integration

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"clims-backend/internal/features/apidocs"
	"clims-backend/internal/features/auth"
	"clims-backend/internal/features/health"
	"clims-backend/internal/features/rbac"
	"clims-backend/internal/features/users"
	"clims-backend/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

func documentationApp() *fiber.App {
	app := fiber.New()
	repo := users.NewRepository(nil)
	service := users.NewService(repo)
	roles := rbac.NewRepository(nil)
	jwt := middleware.NewJWTMiddleware("test-secret")
	active := users.ActiveUser(repo)
	v1 := app.Group("/api/v1")
	auth.RegisterRoutes(v1, auth.NewHandler(repo, service, "test-secret"), jwt, active)
	users.RegisterRoutes(v1, users.NewHandler(repo, service, roles), jwt, active)
	rbac.RegisterRoutes(v1, rbac.NewHandler(roles), jwt, active)
	health.RegisterRoutes(app, health.NewHandler(nil, "test", "test", "test"))
	apidocs.RegisterRoutes(app, os.DirFS(filepath.Join("..", "..", "..", "docs", "api")))
	return app
}

// Check the contract against actual Fiber registration, including health probes.
func TestOpenAPIRouteCoverage(t *testing.T) {
	app := documentationApp()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "api", "openapi.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err = json.Unmarshal(data, &spec); err != nil {
		t.Fatal(err)
	}
	implemented := map[string]bool{}
	for _, route := range app.GetRoutes(true) {
		if route.Method == "HEAD" || route.Method == "OPTIONS" || strings.HasPrefix(route.Path, "/docs") {
			continue
		}
		path := strings.TrimSuffix(route.Path, "/")
		for _, parameter := range route.Params {
			path = strings.ReplaceAll(path, ":"+parameter, "{"+parameter+"}")
		}
		implemented[strings.ToLower(route.Method)+" "+path] = true
	}
	documented := map[string]bool{}
	operationIDs := map[string]bool{}
	for path, methods := range spec.Paths {
		for method, raw := range methods {
			if !strings.Contains(" get post put patch delete head options trace ", " "+method+" ") {
				continue
			}
			key := method + " " + path
			documented[key] = true
			if !implemented[key] {
				t.Errorf("documented operation has no route: %s", key)
			}
			var operation struct {
				ID string `json:"operationId"`
			}
			if err = json.Unmarshal(raw, &operation); err != nil {
				t.Fatal(err)
			}
			if operation.ID == "" || operationIDs[operation.ID] {
				t.Errorf("missing or duplicate operationId for %s", key)
			}
			operationIDs[operation.ID] = true
		}
	}
	for key := range implemented {
		if !documented[key] {
			t.Errorf("route missing from OpenAPI: %s", key)
		}
	}
}

func TestSwaggerAssets(t *testing.T) {
	app := documentationApp()
	for path, contentType := range map[string]string{
		"/docs/":                            "text/html",
		"/docs/openapi.json":                "application/json",
		"/docs/swagger-initializer.js":      "application/javascript",
		"/docs/assets/swagger-ui-bundle.js": "application/javascript",
		"/docs/assets/swagger-ui.css":       "text/css",
	} {
		res, err := app.Test(httptest.NewRequest("GET", path, nil))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != 200 {
			t.Errorf("%s: status %d", path, res.StatusCode)
		}
		if !strings.HasPrefix(res.Header.Get("Content-Type"), contentType) {
			t.Errorf("%s: unexpected content type", path)
		}
	}
	res, err := app.Test(httptest.NewRequest("GET", "/docs", nil))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode < 300 || res.StatusCode > 399 || res.Header.Get("Location") != "/docs/" {
		t.Fatal("missing Swagger directory redirect")
	}
	for _, path := range []string{"/docs/README.md", "/docs/assets/manifest.json", "/docs/../README.md"} {
		res, err := app.Test(httptest.NewRequest("GET", path, nil))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != 404 {
			t.Errorf("unexpected repository exposure at %s", path)
		}
	}
}
