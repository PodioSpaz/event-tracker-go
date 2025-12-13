package main

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/PodioSpaz/event-tracker-go/internal/config"
	"github.com/PodioSpaz/event-tracker-go/internal/db/migrations"
	"github.com/PodioSpaz/event-tracker-go/internal/repository/sqlite"
	"github.com/PodioSpaz/event-tracker-go/internal/ui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Set up logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Set log level
	level, err := zerolog.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	log.Info().
		Str("version", cfg.App.Version).
		Str("name", cfg.App.Name).
		Msg("Event Tracker starting")

	// Ensure database directory exists
	dbPath := cfg.GetDatabasePath()
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatal().Err(err).Str("dir", dbDir).Msg("Failed to create database directory")
	}

	// Open database connection
	db, err := sqlite.New(dbPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Run migrations from embedded files
	log.Info().Msg("Running database migrations")

	// Create temp directory for migrations
	tmpDir, err := os.MkdirTemp("", "event-tracker-migrations-")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create temp directory for migrations")
	}
	defer os.RemoveAll(tmpDir)

	// Extract embedded migrations to temp directory
	migrationsFS := migrations.FS()
	entries, err := fs.ReadDir(migrationsFS, ".")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read embedded migrations")
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		content, err := fs.ReadFile(migrationsFS, entry.Name())
		if err != nil {
			log.Fatal().Err(err).Str("file", entry.Name()).Msg("Failed to read migration file")
		}

		destPath := filepath.Join(tmpDir, entry.Name())
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			log.Fatal().Err(err).Str("file", destPath).Msg("Failed to write migration file")
		}

		log.Debug().Str("file", entry.Name()).Msg("Extracted migration file")
	}

	// Run migrations from temp directory
	if err := db.RunMigrations(tmpDir); err != nil {
		log.Fatal().Err(err).Msg("Failed to run database migrations")
	}

	log.Info().Msg("Database initialized successfully")

	// Create and run GUI application
	app := ui.NewApp(db)
	app.Run()

	log.Info().Msg("Application exiting")
}
