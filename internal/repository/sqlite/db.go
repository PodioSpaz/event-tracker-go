package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

// DB wraps the SQL database connection
type DB struct {
	*sql.DB
}

type txContextKey struct{}

type sqlExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func contextWithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

func txFromContext(ctx context.Context) *sql.Tx {
	if ctx == nil {
		return nil
	}
	tx, _ := ctx.Value(txContextKey{}).(*sql.Tx)
	return tx
}

func (db *DB) getExecutor(ctx context.Context) sqlExecutor {
	if tx := txFromContext(ctx); tx != nil {
		return tx
	}
	return db.DB
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open SQLite database with connection options
	// - Enable foreign keys (required for CASCADE deletes)
	// - Use WAL mode for better concurrency
	// - Set cache size for performance
	dsn := fmt.Sprintf("%s?_foreign_keys=on&_journal_mode=WAL&cache=shared", dbPath)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	// For SQLite, limit to 1 connection to avoid locking issues
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().Str("path", dbPath).Msg("Database connection established")

	return &DB{DB: db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	log.Info().Msg("Closing database connection")
	return db.DB.Close()
}

// RunMigrations executes all SQL migration files
func (db *DB) RunMigrations(migrationsPath string) error {
	log.Info().Str("path", migrationsPath).Msg("Running database migrations")

	// Read migration files
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Execute each migration file in order
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		migrationPath := filepath.Join(migrationsPath, file.Name())
		log.Info().Str("file", file.Name()).Msg("Executing migration")

		// Read migration file
		content, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		// Execute migration
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file.Name(), err)
		}

		log.Info().Str("file", file.Name()).Msg("Migration executed successfully")
	}

	log.Info().Msg("All migrations completed successfully")
	return nil
}

// BeginTx starts a new transaction
func (db *DB) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return db.DB.BeginTx(ctx, nil)
}

// HealthCheck verifies the database connection is healthy
func (db *DB) HealthCheck(ctx context.Context) error {
	return db.PingContext(ctx)
}

// GetStats returns database connection statistics
func (db *DB) GetStats() sql.DBStats {
	return db.DB.Stats()
}

// WithTransaction executes a function within a transaction
// If the function returns an error, the transaction is rolled back.
// Nested transactions reuse the existing transaction in the context.
func (db *DB) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if existingTx := txFromContext(ctx); existingTx != nil {
		return fn(ctx)
	}

	tx, err := db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback on panic
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Re-throw panic after rollback
		}
	}()

	// Execute the function
	txCtx := contextWithTx(ctx, tx)
	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Error().Err(rbErr).Msg("Failed to rollback transaction")
		}
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ExecContext executes the query using the transaction in ctx if one exists.
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.getExecutor(ctx).ExecContext(ctx, query, args...)
}

// QueryContext executes the query using the transaction in ctx if one exists.
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.getExecutor(ctx).QueryContext(ctx, query, args...)
}

// QueryRowContext executes the query using the transaction in ctx if one exists.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.getExecutor(ctx).QueryRowContext(ctx, query, args...)
}
