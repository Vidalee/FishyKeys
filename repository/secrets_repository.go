package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5"
	"time"

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
	ListSecrets(ctx context.Context) ([]Secret, error)
	DeleteSecret(ctx context.Context, path string) error
	HasAccess(ctx context.Context, secretPath string, userID *int, roleIDs []int) (bool, error)
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

func (r *secretsRepository) ListSecrets(ctx context.Context) ([]Secret, error) {
	query := `SELECT id, path, encrypted_encryption_key, encrypted_value, owner_user_id, created_at, updated_at FROM secrets ORDER BY created_at`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secrets []Secret
	for rows.Next() {
		var s Secret
		if err := rows.Scan(
			&s.ID,
			&s.Path,
			&s.EncryptedEncryptionKey,
			&s.EncryptedValue,
			&s.OwnerUserId,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		secrets = append(secrets, s)
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
