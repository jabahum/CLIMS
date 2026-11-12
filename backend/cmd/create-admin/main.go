// create-admin provisions an administrator using an operator's database access.
package main

import (
	"context"
	"log"
	"os"
	"strings"

	"clims-backend/internal/config"
	"clims-backend/internal/database"
	"clims-backend/internal/features/users"
	"clims-backend/internal/pkg/validator"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	req := users.AccountInput{
		Username: strings.TrimSpace(os.Getenv("CLIMS_ADMIN_USERNAME")),
		Email:    strings.ToLower(strings.TrimSpace(os.Getenv("CLIMS_ADMIN_EMAIL"))),
		Password: os.Getenv("CLIMS_ADMIN_PASSWORD"),
	}
	if req.Username == "" {
		req.Username = "admin"
	}
	if err := validator.Validate.Struct(req); err != nil {
		log.Fatal("provide valid CLIMS_ADMIN_EMAIL, CLIMS_ADMIN_PASSWORD (8–72 bytes), and optional CLIMS_ADMIN_USERNAME")
	}
	if len(req.Password) > 72 {
		log.Fatal("CLIMS_ADMIN_PASSWORD must not exceed 72 bytes")
	}
	if err := database.RunMigrations(cfg.DSN()); err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	repo := users.NewRepository(pool)
	id, err := users.NewService(repo).Create(ctx, req, 1)
	if err != nil {
		log.Fatal("failed to create admin; ensure its email is unused")
	}
	log.Printf("Admin account created with user ID %d", id)
}
