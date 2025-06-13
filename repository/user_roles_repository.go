package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRolesRepository interface {
	AssignRoleToUser(ctx context.Context, userID, roleID int) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID int) error
	GetUserRoleIDs(ctx context.Context, userID int) ([]int, error)
}

type userRolesRepository struct {
	pool *pgxpool.Pool
}

func NewUserRolesRepository(pool *pgxpool.Pool) UserRolesRepository {
	return &userRolesRepository{pool: pool}
}

func (r *userRolesRepository) AssignRoleToUser(ctx context.Context, userID, roleID int) error {
	query := `
		INSERT INTO user_roles (user_id, role_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, userID, roleID, time.Now().UTC())
	return err
}

func (r *userRolesRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	_, err := r.pool.Exec(ctx, query, userID, roleID)
	return err
}

func (r *userRolesRepository) GetUserRoleIDs(ctx context.Context, userID int) ([]int, error) {
	query := `SELECT role_id FROM user_roles WHERE user_id = $1`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roleIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		roleIDs = append(roleIDs, id)
	}
	return roleIDs, nil
}
