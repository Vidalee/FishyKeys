package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Vidalee/FishyKeys/internal/crypto"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSecretNotFound = errors.New("secret not found")
)

type Secret struct {
	ID                     int
	Path                   string
	EncryptedEncryptionKey string
	EncryptedValue         string
	OwnerUserId            int
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type DecryptedSecret struct {
	ID                     int
	Path                   string
	DecryptedEncryptionKey string
	DecryptedValue         string
	OwnerUserId            int
	AuthorizedUserIDs      []int
	AuthorizedRoleIDs      []int
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type SecretsRepository interface {
	CreateSecret(ctx context.Context, keyManager *crypto.KeyManager, path string, ownerUserId int, value string) (int, error)
	GetSecretByPath(ctx context.Context, keyManager *crypto.KeyManager, path string) (*DecryptedSecret, error)
	ListSecretsForUser(ctx context.Context, userID int) ([]Secret, error)
	DeleteSecret(ctx context.Context, path string) error
	HasAccess(ctx context.Context, secretPath string, userID *int, roleIDs []int) (bool, error)
	UpdateSecret(ctx context.Context, keyManager *crypto.KeyManager, path string, newValue string) error
}

type secretsRepository struct {
	pool *pgxpool.Pool
}

func NewSecretsRepository(pool *pgxpool.Pool) SecretsRepository {
	return &secretsRepository{pool: pool}
}

func (r *secretsRepository) CreateSecret(ctx context.Context, keyManager *crypto.KeyManager, path string, ownerUserId int, value string) (int, error) {
	encryptionKey, err := crypto.GenerateSecret()
	if err != nil {
		return 0, err
	}

	encryptedValue, err := crypto.EncryptWithKey(encryptionKey, []byte(value))
	if err != nil {
		return 0, err
	}

	encryptedEncryptionKey, err := crypto.Encrypt(keyManager, encryptionKey)
	if err != nil {
		return 0, err
	}

	query := `
INSERT INTO secrets (path, encrypted_encryption_key, encrypted_value, owner_user_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $5)
ON CONFLICT (path) DO NOTHING
RETURNING id
`
	var secretID int
	now := time.Now().UTC()
	err = r.pool.QueryRow(ctx, query, path, encryptedEncryptionKey, encryptedValue, ownerUserId, now).Scan(&secretID)
	if err != nil {
		return 0, err
	}
	return secretID, nil
}

func (r *secretsRepository) GetSecretByPath(ctx context.Context, keyManager *crypto.KeyManager, path string) (*DecryptedSecret, error) {
	query := `SELECT id, path, encrypted_encryption_key, encrypted_value, owner_user_id, created_at, updated_at FROM secrets WHERE path = $1`
	var secret Secret
	err := r.pool.QueryRow(ctx, query, path).Scan(
		&secret.ID,
		&secret.Path,
		&secret.EncryptedEncryptionKey,
		&secret.EncryptedValue,
		&secret.OwnerUserId,
		&secret.CreatedAt,
		&secret.UpdatedAt,
	)
	if err != nil {
		return nil, ErrSecretNotFound
	}

	decryptedKey, err := crypto.Decrypt(keyManager, secret.EncryptedEncryptionKey)
	if err != nil {
		return nil, err
	}
	decryptedValue, err := crypto.DecryptWithKey(decryptedKey, secret.EncryptedValue)
	if err != nil {
		return nil, err
	}

	var authorizedUserIDs []int
	var authorizedRoleIDs []int
	accessQuery := `
SELECT user_id, role_id
FROM secrets_access
WHERE secret_id = $1
`
	rows, err := r.pool.Query(ctx, accessQuery, secret.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var userID, roleID sql.NullInt64
		if err := rows.Scan(&userID, &roleID); err != nil {
			return nil, err
		}
		if userID.Valid {
			authorizedUserIDs = append(authorizedUserIDs, int(userID.Int64))
		}
		if roleID.Valid {
			authorizedRoleIDs = append(authorizedRoleIDs, int(roleID.Int64))
		}
	}

	return &DecryptedSecret{
		ID:                     secret.ID,
		Path:                   secret.Path,
		DecryptedEncryptionKey: string(decryptedKey),
		DecryptedValue:         string(decryptedValue),
		OwnerUserId:            secret.OwnerUserId,
		AuthorizedUserIDs:      authorizedUserIDs,
		AuthorizedRoleIDs:      authorizedRoleIDs,
		CreatedAt:              secret.CreatedAt,
		UpdatedAt:              secret.UpdatedAt,
	}, nil
}

func (r *secretsRepository) ListSecretsForUser(ctx context.Context, userID int) ([]Secret, error) {
	// Get the user's role IDs
	roleIDs := []int{}
	roleRows, err := r.pool.Query(ctx, `SELECT role_id FROM user_roles WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer roleRows.Close()
	for roleRows.Next() {
		var id int
		if err := roleRows.Scan(&id); err != nil {
			return nil, err
		}
		roleIDs = append(roleIDs, id)
	}

	query := `
SELECT DISTINCT ON (s.id) s.id, s.path, s.encrypted_encryption_key, s.encrypted_value, s.owner_user_id, s.created_at, s.updated_at
FROM secrets s
LEFT JOIN secrets_access sa ON sa.secret_id = s.id
WHERE s.owner_user_id = $1 OR sa.user_id = $1 OR sa.role_id = ANY($2)
ORDER BY s.id`

	rows, err := r.pool.Query(ctx, query, userID, roleIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secrets []Secret
	for rows.Next() {
		var secret Secret
		if err := rows.Scan(
			&secret.ID,
			&secret.Path,
			&secret.EncryptedEncryptionKey,
			&secret.EncryptedValue,
			&secret.OwnerUserId,
			&secret.CreatedAt,
			&secret.UpdatedAt,
		); err != nil {
			return nil, err
		}
		secrets = append(secrets, secret)
	}
	return secrets, nil
}

func (r *secretsRepository) DeleteSecret(ctx context.Context, path string) error {
	cmd, err := r.pool.Exec(ctx, `DELETE FROM secrets WHERE path = $1`, path)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrSecretNotFound
	}
	return nil
}

func (r *secretsRepository) HasAccess(ctx context.Context, secretPath string, userID *int, roleIDs []int) (bool, error) {
	if secretPath == "" {
		return false, errors.New("secretPath must not be empty")
	}

	query := `
SELECT 1
FROM secrets s
LEFT JOIN secrets_access sa ON sa.secret_id = s.id
WHERE s.path = $1
AND (
    s.owner_user_id = $2 OR
    sa.user_id = $2 OR
    sa.role_id = ANY($3)
)
LIMIT 1
`

	var exists int
	err := r.pool.QueryRow(ctx, query, secretPath, *userID, roleIDs).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *secretsRepository) UpdateSecret(ctx context.Context, keyManager *crypto.KeyManager, path string, newValue string) error {
	encryptionKey, err := crypto.GenerateSecret()
	if err != nil {
		return err
	}

	encryptedValue, err := crypto.EncryptWithKey(encryptionKey, []byte(newValue))
	if err != nil {
		return err
	}

	encryptedEncryptionKey, err := crypto.Encrypt(keyManager, encryptionKey)
	if err != nil {
		return err
	}

	query := `
UPDATE secrets
SET encrypted_encryption_key = $1,
    encrypted_value = $2,
    updated_at = $3
WHERE path = $4
`
	cmd, err := r.pool.Exec(ctx, query, encryptedEncryptionKey, encryptedValue, time.Now().UTC(), path)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrSecretNotFound
	}
	return nil
}
