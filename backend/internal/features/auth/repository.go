package auth

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
}

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) CreateUser(ctx context.Context, username, email, passwordHash string) (int, error) {
	var userID int
	query := `INSERT INTO users (username, email, password, firstname, lastname, role, phone_number, status) VALUES ($1, $2, $3, '', '', 'user', '', 'active') RETURNING id`
	err := r.DB.QueryRow(ctx, query, username, email, passwordHash).Scan(&userID)
	return userID, err
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	query := `SELECT id, username, email, password FROM users WHERE email = $1`
	err := r.DB.QueryRow(ctx, query, email).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID int) (*User, error) {
	var u User
	query := `SELECT id, username, email, password FROM users WHERE id = $1`
	err := r.DB.QueryRow(ctx, query, userID).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
