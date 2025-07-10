package repository

import (
	"context"
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
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type DecryptedSecret struct {
	ID                     int
	Path                   string
	DecryptedEncryptionKey string
	DecryptedValue         string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type SecretsRepository interface {
	CreateSecret(ctx context.Context, keyManager *crypto.KeyManager, path string, value string) (int, error)
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

func (r *secretsRepository) CreateSecret(ctx context.Context, keyManager *crypto.KeyManager, path string, value string) (int, error) {
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
INSERT INTO secrets (path, encrypted_encryption_key, encrypted_value, created_at, updated_at)
VALUES ($1, $2, $3, $4, $4)
ON CONFLICT (path) DO UPDATE SET 
    encrypted_encryption_key = EXCLUDED.encrypted_encryption_key,
    encrypted_value = EXCLUDED.encrypted_value,
    updated_at = EXCLUDED.updated_at
RETURNING id
`
	var secretID int
	now := time.Now().UTC()
	err = r.pool.QueryRow(ctx, query, path, encryptedEncryptionKey, encryptedValue, now).Scan(&secretID)
	if err != nil {
		return 0, err
	}
	return secretID, nil
}

func (r *secretsRepository) GetSecretByPath(ctx context.Context, keyManager *crypto.KeyManager, path string) (*DecryptedSecret, error) {
	query := `SELECT id, path, encrypted_encryption_key, encrypted_value, created_at, updated_at FROM secrets WHERE path = $1`
	var secret Secret
	err := r.pool.QueryRow(ctx, query, path).Scan(
		&secret.ID,
		&secret.Path,
		&secret.EncryptedEncryptionKey,
		&secret.EncryptedValue,
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
	return &DecryptedSecret{
		ID:                     secret.ID,
		Path:                   secret.Path,
		DecryptedEncryptionKey: string(decryptedKey),
		DecryptedValue:         string(decryptedValue),
		CreatedAt:              secret.CreatedAt,
		UpdatedAt:              secret.UpdatedAt,
	}, nil
}

func (r *secretsRepository) ListSecrets(ctx context.Context) ([]Secret, error) {
	query := `SELECT id, path, encrypted_encryption_key, encrypted_value, created_at, updated_at FROM secrets ORDER BY created_at`
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
	if userID == nil || roleIDs == nil {
		return false, errors.New("must provide either userID and roleIDs (can be empty) for access check")
	}

	query := `
SELECT 1
FROM secrets_access sa
JOIN secrets s ON sa.secret_id = s.id
WHERE s.path = $1
AND (
    (sa.user_id = $2) OR
    (sa.role_id = ANY($3))
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
