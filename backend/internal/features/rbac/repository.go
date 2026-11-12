package rbac

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAdminRole = errors.New("admin permissions cannot be changed")

type Repository struct{ DB *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{DB: db} }
func (r *Repository) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := r.DB.Query(ctx, `SELECT r.id,r.name,ARRAY(SELECT permission_code FROM role_permissions WHERE role_id=r.id ORDER BY permission_code) FROM roles r ORDER BY r.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roles := []Role{}
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Permissions); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *Repository) RoleExists(ctx context.Context, id int) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM roles WHERE id=$1)`, id).Scan(&exists)
	return exists, err
}

func (r *Repository) ListPermissions(ctx context.Context) ([]string, error) {
	rows, err := r.DB.Query(ctx, `SELECT code FROM permissions ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	codes := []string{}
	for rows.Next() {
		var code string
		if err = rows.Scan(&code); err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	return codes, rows.Err()
}

func (r *Repository) SetRolePermissions(ctx context.Context, id int, codes []string) error {
	if id == 1 {
		return ErrAdminRole
	}
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var roleID int
	if err = tx.QueryRow(ctx, `SELECT id FROM roles WHERE id=$1 FOR UPDATE`, id).Scan(&roleID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM role_permissions WHERE role_id=$1`, id); err != nil {
		return err
	}
	for _, code := range codes {
		if _, err = tx.Exec(ctx, `INSERT INTO role_permissions(role_id,permission_code) VALUES($1,$2) ON CONFLICT DO NOTHING`, id, code); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
