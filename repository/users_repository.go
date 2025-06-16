package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type User struct {
	ID        int
	Username  string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UsersRepository interface {
	CreateUser(ctx context.Context, username string, hashedPassword string) (int, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	ListUsers(ctx context.Context) ([]User, error)
	DeleteUser(ctx context.Context, username string) error
}

type usersRepository struct {
	pool *pgxpool.Pool
}

func NewUsersRepository(pool *pgxpool.Pool) UsersRepository {
	return &usersRepository{pool: pool}
}

func (r *usersRepository) CreateUser(ctx context.Context, username string, hashedPassword string) (int, error) {
	query := `
INSERT INTO users (username, password, created_at, updated_at)
VALUES ($1, $2, $3, $3)
ON CONFLICT (username) DO NOTHING
RETURNING id
`
	var userID int
	err := r.pool.QueryRow(ctx, query, username, hashedPassword, time.Now().UTC()).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *usersRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `SELECT id, username, password, created_at, updated_at FROM users WHERE username = $1`
	var user User
	err := r.pool.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

func (r *usersRepository) ListUsers(ctx context.Context) ([]User, error) {
	query := `SELECT id, username, password, created_at, updated_at FROM users ORDER BY created_at`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Password, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *usersRepository) DeleteUser(ctx context.Context, username string) error {
	cmd, err := r.pool.Exec(ctx, `DELETE FROM users WHERE username = $1`, username)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}
