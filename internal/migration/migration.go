package migration

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// RunMigrations runs all pending migrations
func RunMigrations(pool *pgxpool.Pool) error {
	config := pool.Config()
	connString := config.ConnString()

	db, err := sql.Open("pgx", connString)
	if err != nil {
		return fmt.Errorf("could not create sql.DB: %v", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %v", err)
	}

	// Not very cool hotfix for not having to change the working directory
	// when running tests individually in GoLand...
	migrationPath := "file://migrations"
	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		if _, err := os.Stat("../migrations"); err == nil {
			migrationPath = "file://" + filepath.ToSlash("../migrations")
		}
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migration instance: %v", err)
	}

	_, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("could not check migration version: %v", err)
	}

	if errors.Is(err, migrate.ErrNilVersion) || dirty {
		log.Println("Detected migrations to be applied")
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("could not run migrations: %v", err)
	}

	// Only print success message if we actually ran migrations
	if errors.Is(err, migrate.ErrNilVersion) || dirty {
		log.Println("Database migrations completed successfully")
	}
	return nil
}
