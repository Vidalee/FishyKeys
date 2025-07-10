package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SecretsAccessRepository interface {
	GrantUserAccess(ctx context.Context, secretPath string, userID int) error
	GrantRoleAccess(ctx context.Context, secretPath string, roleID int) error
	RevokeUserAccess(ctx context.Context, secretPath string, userID int) error
	RevokeRoleAccess(ctx context.Context, secretPath string, roleID int) error
}

type secretsAccessRepository struct {
	pool *pgxpool.Pool
}

func NewSecretsAccessRepository(pool *pgxpool.Pool) SecretsAccessRepository {
	return &secretsAccessRepository{pool: pool}
}

func (r *secretsAccessRepository) GrantUserAccess(ctx context.Context, secretPath string, userID int) error {
	if secretPath == "" {
		return errors.New("secretPath must not be empty")
	}

	var secretID int
	err := r.pool.QueryRow(ctx, `SELECT id FROM secrets WHERE path = $1`, secretPath).Scan(&secretID)
	if err != nil {
		return ErrSecretNotFound
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO secrets_access (secret_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, secretID, userID)
	return err
}

func (r *secretsAccessRepository) GrantRoleAccess(ctx context.Context, secretPath string, roleID int) error {
	if secretPath == "" {
		return errors.New("secretPath must not be empty")
	}

	var secretID int
	err := r.pool.QueryRow(ctx, `SELECT id FROM secrets WHERE path = $1`, secretPath).Scan(&secretID)
	if err != nil {
		return ErrSecretNotFound
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO secrets_access (secret_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, secretID, roleID)
	return err
}

func (r *secretsAccessRepository) RevokeUserAccess(ctx context.Context, secretPath string, userID int) error {
	if secretPath == "" {
		return errors.New("secretPath must not be empty")
	}

	_, err := r.pool.Exec(ctx, `
		DELETE FROM secrets_access
		WHERE secret_id = (SELECT id FROM secrets WHERE path = $1)
		AND user_id = $2
	`, secretPath, userID)
	return err
}

func (r *secretsAccessRepository) RevokeRoleAccess(ctx context.Context, secretPath string, roleID int) error {
	if secretPath == "" {
		return errors.New("secretPath must not be empty")
	}

	_, err := r.pool.Exec(ctx, `
		DELETE FROM secrets_access
		WHERE secret_id = (SELECT id FROM secrets WHERE path = $1)
		AND role_id = $2
	`, secretPath, roleID)
	return err
}
