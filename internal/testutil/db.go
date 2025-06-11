package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/Vidalee/FishyKeys/internal/migration"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testContainer testcontainers.Container
	testDB        *pgxpool.Pool
)

func SetupTestDB() (*pgxpool.Pool, error) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort("5432/tcp"),
		),
	}

	var err error
	testContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	port, err := testContainer.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	host, err := testContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	dsn := fmt.Sprintf("postgres://test:test@%s:%d/testdb?sslmode=disable", host, port.Int())
	testDB, err = pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	for i := 0; i < 100; i++ {
		if err := testDB.Ping(ctx); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
		if i == 99 {
			return nil, fmt.Errorf("database not ready after 10 seconds")
		}
	}

	if err := migration.RunMigrations(testDB); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return testDB, nil
}

func TeardownTestDB() error {
	if testContainer != nil {
		if err := testContainer.Terminate(context.Background()); err != nil {
			return fmt.Errorf("failed to terminate container: %w", err)
		}
	}
	if testDB != nil {
		testDB.Close()
	}
	return nil
}

func ClearTable(ctx context.Context, tableName string) error {
	_, err := testDB.Exec(ctx, fmt.Sprintf("DELETE FROM %s", tableName))
	if err != nil {
		return fmt.Errorf("failed to clear table %s: %w", tableName, err)
	}
	return nil
}
