package main

import (
	"context"
	"log"

	"clims-backend/internal/config"
	"clims-backend/internal/database"
	"clims-backend/internal/features/auth"
	"clims-backend/internal/features/health"
	"clims-backend/internal/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Version variables injected at build time via -ldflags
var (
	Version    = "dev"
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
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Register Health Probes with Version info
	healthHandler := health.NewHandler(dbPool, Version, CommitHash, BuildTime)
	health.RegisterRoutes(app, healthHandler)

	v1 := app.Group("/api/v1")
	jwtAuthMiddleware := middleware.NewJWTMiddleware(cfg.JWTSecret)

	authRepo := auth.NewRepository(dbPool)
	authHandler := auth.NewHandler(authRepo, cfg.JWTSecret)
	auth.RegisterRoutes(v1, authHandler, jwtAuthMiddleware)

	log.Printf("Server starting on port %s (Version: %s, Commit: %s)...", cfg.Port, Version, CommitHash)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
