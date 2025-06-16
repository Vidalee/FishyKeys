package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRoleNotFound = errors.New("role not found")
)

type Role struct {
	ID        int
	Name      string
	Color     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RolesRepository interface {
	CreateRole(ctx context.Context, name string, color string) error
	GetRoleByID(ctx context.Context, id int) (*Role, error)
	GetRoleByName(ctx context.Context, name string) (*Role, error)
	ListRoles(ctx context.Context) ([]Role, error)
}

type rolesRepository struct {
	pool *pgxpool.Pool
}

func NewRolesRepository(pool *pgxpool.Pool) RolesRepository {
	return &rolesRepository{pool: pool}
}

func (r *rolesRepository) CreateRole(ctx context.Context, name string, color string) error {
	query := `
		INSERT INTO roles (name, color, created_at, updated_at)
		VALUES ($1, $2, $3, $3)
		ON CONFLICT (name) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, name, color, time.Now().UTC())
	return err
}

func (r *rolesRepository) GetRoleByID(ctx context.Context, id int) (*Role, error) {
	query := `SELECT id, name, color, created_at, updated_at FROM roles WHERE id = $1`
	var role Role
	err := r.pool.QueryRow(ctx, query, id).Scan(&role.ID, &role.Name, &role.Color, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	return &role, nil
}

func (r *rolesRepository) GetRoleByName(ctx context.Context, name string) (*Role, error) {
	query := `SELECT id, name, color, created_at, updated_at FROM roles WHERE name = $1`
	var role Role
	err := r.pool.QueryRow(ctx, query, name).Scan(&role.ID, &role.Name, &role.Color, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	return &role, nil
}

func (r *rolesRepository) ListRoles(ctx context.Context) ([]Role, error) {
	query := `SELECT id, name, color, created_at, updated_at FROM roles`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Color, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}
