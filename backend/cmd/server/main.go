package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"clims-backend/internal/config"
	"clims-backend/internal/database"
	"clims-backend/internal/features/apidocs"
	"clims-backend/internal/features/auth"
	"clims-backend/internal/features/health"
	"clims-backend/internal/features/rbac"
	"clims-backend/internal/features/users"
	"clims-backend/internal/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Version variables injected at build time via -ldflags
var (
	Version    = "2.0.0"
	CommitHash = "none"
	BuildTime  = "unknown"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := database.RunMigrations(cfg.DSN()); err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to create connection pool: %v", err)
	}
	defer dbPool.Close()

	app := fiber.New()
	app.Use(recover.New())
	app.Use(middleware.NewSlogLogger())

	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	docsDir := os.Getenv("DOCS_DIR")
	if docsDir == "" {
		docsDir = "../docs"
	}
	apidocs.RegisterRoutes(app, os.DirFS(filepath.Join(docsDir, "api")))

	// Register Health Probes with Version info
	healthHandler := health.NewHandler(dbPool, Version, CommitHash, BuildTime)
	health.RegisterRoutes(app, healthHandler)

	v1 := app.Group("/api/v1")
	jwtAuthMiddleware := middleware.NewJWTMiddleware(cfg.JWTSecret)

	userRepo := users.NewRepository(dbPool)
	userService := users.NewService(userRepo)
	roleRepo := rbac.NewRepository(dbPool)
	activeUser := users.ActiveUser(userRepo)

	auth.RegisterRoutes(v1, auth.NewHandler(userRepo, userService, cfg.JWTSecret), jwtAuthMiddleware, activeUser)
	users.RegisterRoutes(v1, users.NewHandler(userRepo, userService, roleRepo), jwtAuthMiddleware, activeUser)
	rbac.RegisterRoutes(v1, rbac.NewHandler(roleRepo), jwtAuthMiddleware, activeUser)

	log.Printf("Server starting on port %s (Version: %s, Commit: %s)...", cfg.Port, Version, CommitHash)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
