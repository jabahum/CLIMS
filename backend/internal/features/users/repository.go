package users

import (
	"context"
	"errors"
	"strconv"

	"clims-backend/internal/features/rbac"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrLastAdmin = errors.New("cannot disable or demote the last active admin")

type Repository struct{ DB *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{DB: db} }
func (r *Repository) CreateUser(ctx context.Context, req AccountInput, passwordHash string, roleID int) (int, error) {
	var id int
	role := ""
	if roleID > 0 {
		role = strconv.Itoa(roleID)
	}
	err := r.DB.QueryRow(ctx, `INSERT INTO users
		(username,email,password,firstname,lastname,phone_number,role,status,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP) RETURNING id`,
		req.Username, req.Email, passwordHash, req.Firstname, req.Lastname, req.Phone, role).Scan(&id)
	return id, err
}

const userSelect = `SELECT u.id,u.username,COALESCE(u.email,''),u.password,u.firstname,u.lastname,u.phone_number,u.status,
	r.id,r.name,ARRAY(SELECT permission_code FROM role_permissions WHERE role_id=r.id ORDER BY permission_code)
	FROM users u LEFT JOIN roles r ON r.id::text=u.role `

func scanUser(row pgx.Row) (*User, error) {
	var u User
	var roleID *int
	var roleName *string
	var permissions []string
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Firstname, &u.Lastname, &u.Phone, &u.Status, &roleID, &roleName, &permissions)
	if err != nil {
		return nil, err
	}
	if roleID != nil {
		u.Role = &rbac.Role{ID: *roleID, Name: *roleName, Permissions: permissions}
	}
	return &u, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return scanUser(r.DB.QueryRow(ctx, userSelect+`WHERE lower(btrim(u.email))=lower(btrim($1))`, email))
}

func (r *Repository) GetUserByID(ctx context.Context, id int) (*User, error) {
	return scanUser(r.DB.QueryRow(ctx, userSelect+`WHERE u.id=$1`, id))
}

func (r *Repository) ListUsers(ctx context.Context, limit, offset int) ([]User, error) {
	rows, err := r.DB.Query(ctx, userSelect+`ORDER BY u.id LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}

// Serialize account changes so concurrent requests cannot remove every active admin.
func (r *Repository) UpdateUserAccess(ctx context.Context, id int, roleID *int, status *string) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(68421301)`); err != nil {
		return err
	}
	var currentRole, currentStatus string
	if err = tx.QueryRow(ctx, `SELECT role,status FROM users WHERE id=$1 FOR UPDATE`, id).Scan(&currentRole, &currentStatus); err != nil {
		return err
	}
	role, newStatus := currentRole, currentStatus
	if roleID != nil {
		role = strconv.Itoa(*roleID)
	}
	if status != nil {
		newStatus = *status
	}
	if currentRole == "1" && currentStatus == "active" && (role != "1" || newStatus != "active") {
		var admins int
		if err = tx.QueryRow(ctx, `SELECT count(*) FROM users WHERE role='1' AND status='active'`).Scan(&admins); err != nil {
			return err
		}
		if admins <= 1 {
			return ErrLastAdmin
		}
	}
	if roleID != nil {
		var exists bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM roles WHERE id=$1)`, *roleID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return pgx.ErrNoRows
		}
	}
	_, err = tx.Exec(ctx, `UPDATE users SET role=$2,status=$3,updated_at=CURRENT_TIMESTAMP WHERE id=$1`, id, role, newStatus)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
