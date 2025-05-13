package migrator

import (
	"database/sql"
	"fmt"
	"net/url"

	_ "github.com/go-sql-driver/mysql" // MySQL database/sql driver
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql" // MySQL driver
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Migrator struct {
	m *migrate.Migrate
}

// NewMigrator creates a new database migrator instance.
// It takes a migrations path and database URL as input.
// Returns a Migrator instance and error if any step fails.
func NewMigrator(migrationsPath, databaseURL string) (*Migrator, error) {
	if _, err := url.Parse(databaseURL); err != nil {
		return nil, fmt.Errorf("invalid database URL: %w", err)
	}

	db, err := sql.Open("mysql", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MySQL driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"mysql",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrator: %w", err)
	}

	return &Migrator{m: m}, nil
}

// Close closes the migrator instance and releases associated resources.
func (m *Migrator) Close() {
	if m.m != nil {
		m.m.Close()
	}
}

// NewMySQLDSN creates a MySQL connection string (DSN) from individual connection parameters.
// Parameters:
//   - user: database username
//   - password: database password
//   - host: database host
//   - port: database port
//   - dbName: database name
//
// Returns formatted DSN string with proper escaping and connection options.
func NewMySQLDSN(user, password, host, port, dbName string) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&multiStatements=true",
		url.PathEscape(user),
		url.PathEscape(password),
		host,
		port,
		dbName,
	)
}

// Up applies all available up migrations.
// Returns error if migration fails, nil if successful or no changes needed.
func (m *Migrator) Up() error {
	if err := m.m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("up migration failed: %w", err)
	}
	return nil
}

// Down rolls back all migrations.
// Returns error if migration fails, nil if successful or no changes needed.
func (m *Migrator) Down() error {
	if err := m.m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("down migration failed: %w", err)
	}
	return nil
}

// Steps applies n migrations in the direction specified by n.
// n > 0 means up migrations, n < 0 means down migrations.
// Returns error if migration fails, nil if successful or no changes needed.
func (m *Migrator) Steps(steps int) error {
	if err := m.m.Steps(steps); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("step migration failed: %w", err)
	}
	return nil
}

// Version returns the currently active migration version.
// Returns the version number, whether a dirty state was detected, and error if any.
func (m *Migrator) Version() (uint, bool, error) {
	return m.m.Version()
}
