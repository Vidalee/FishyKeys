package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSettingNotFound = errors.New("setting not found")
)

// GlobalSettingsRepository defines the interface for global settings storage
type GlobalSettingsRepository interface {
	// StoreSetting stores a string value for a given key
	StoreSetting(ctx context.Context, key string, value string) error
	// StoreSettings stores multiple key-value pairs in a single query
	StoreSettings(ctx context.Context, settings map[string]string) error
	// GetSetting retrieves a string value for a given key
	GetSetting(ctx context.Context, key string) (string, error)
	// GetSettings retrieves values for multiple keys
	GetSettings(ctx context.Context, keys ...string) (map[string]string, error)
	// DeleteSetting deletes a single setting by its key
	DeleteSetting(ctx context.Context, key string) error
	// DeleteSettings deletes multiple settings by their keys
	DeleteSettings(ctx context.Context, keys ...string) error
}

type globalSettingsRepository struct {
	pool *pgxpool.Pool
}

func NewGlobalSettingsRepository(pool *pgxpool.Pool) GlobalSettingsRepository {
	return &globalSettingsRepository{pool: pool}
}

func (r *globalSettingsRepository) StoreSetting(ctx context.Context, key string, value string) error {
	query := `
		INSERT INTO global_settings (key, value, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (key) DO UPDATE
		SET value = $2, updated_at = $3
	`

	_, err := r.pool.Exec(ctx, query, key, value, time.Now().UTC())
	return err
}

func (r *globalSettingsRepository) StoreSettings(ctx context.Context, settings map[string]string) error {
	if len(settings) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(ctx, "store_settings", `
		INSERT INTO global_settings (key, value, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (key) DO UPDATE
		SET value = $2, updated_at = $3
	`)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	for key, value := range settings {
		_, err = tx.Exec(ctx, stmt.Name, key, value, now)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *globalSettingsRepository) GetSetting(ctx context.Context, key string) (string, error) {
	var value string

	query := `
		SELECT value
		FROM global_settings
		WHERE key = $1
	`

	err := r.pool.QueryRow(ctx, query, key).Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrSettingNotFound
		}
		return "", err
	}

	return value, nil
}

func (r *globalSettingsRepository) GetSettings(ctx context.Context, keys ...string) (map[string]string, error) {
	if len(keys) == 0 {
		return make(map[string]string), nil
	}

	query := `
		SELECT key, value
		FROM global_settings
		WHERE key = ANY($1)
	`

	rows, err := r.pool.Query(ctx, query, keys)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(result) != len(keys) {
		return nil, ErrSettingNotFound
	}

	return result, nil
}

// DeleteSetting deletes a single setting by its key
func (r *globalSettingsRepository) DeleteSetting(ctx context.Context, key string) error {
	query := `
		DELETE FROM global_settings
		WHERE key = $1
	`

	cmd, err := r.pool.Exec(ctx, query, key)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrSettingNotFound
	}

	return nil
}

// DeleteSettings deletes multiple settings by their keys
func (r *globalSettingsRepository) DeleteSettings(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	query := `
		DELETE FROM global_settings
		WHERE key = ANY($1)
	`

	cmd, err := r.pool.Exec(ctx, query, keys)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrSettingNotFound
	}

	return nil
}
