package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PodioSpaz/event-tracker-go/internal/config"
	"github.com/PodioSpaz/event-tracker-go/internal/repository/sqlite"
	"github.com/PodioSpaz/event-tracker-go/pkg/migration"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	cmdUp     = "up"
	cmdImport = "import"
	cmdStats  = "stats"
)

func main() {
	// Command line flags
	cmdFlag := flag.String("cmd", cmdUp, "Command to run: up (run SQL migrations), import (import TinyDB data), stats (show database stats)")
	sourceFlag := flag.String("source", "", "Source TinyDB JSON file path (required for import command)")
	dbFlag := flag.String("db", "", "Database path (overrides config)")
	verboseFlag := flag.Bool("verbose", false, "Enable verbose logging")

	flag.Parse()

	// Setup logging
	setupLogging(*verboseFlag)

	// Validate command
	cmd := *cmdFlag
	if cmd != cmdUp && cmd != cmdImport && cmd != cmdStats {
		log.Fatal().Msgf("Invalid command: %s. Must be 'up', 'import', or 'stats'", cmd)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Determine database path
	dbPath := *dbFlag
	if dbPath == "" {
		dbPath = cfg.Database.Path
	}

	// Ensure database directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatal().Err(err).Str("dir", dbDir).Msg("Failed to create database directory")
	}

	log.Info().Str("database", dbPath).Msg("Migration tool started")

	// Open database connection
	db, err := sqlite.New(dbPath)
	if err != nil {
		log.Fatal().Err(err).Str("path", dbPath).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Execute command
	ctx := context.Background()

	switch cmd {
	case cmdUp:
		if err := runSQLMigrations(db); err != nil {
			log.Fatal().Err(err).Msg("SQL migrations failed")
		}
	case cmdImport:
		// Ensure SQL migrations are run first
		if err := runSQLMigrations(db); err != nil {
			log.Fatal().Err(err).Msg("SQL migrations failed")
		}
		if err := runImport(ctx, db, *sourceFlag); err != nil {
			log.Fatal().Err(err).Msg("Import failed")
		}
	case cmdStats:
		if err := runStats(ctx, db); err != nil {
			log.Fatal().Err(err).Msg("Failed to get stats")
		}
	}

	log.Info().Msg("Migration tool completed successfully")
}

// runSQLMigrations executes SQL migration files
func runSQLMigrations(db *sqlite.DB) error {
	log.Info().Msg("Running database migrations")

	// Get migrations directory
	migrationsPath := "migrations"
	if !filepath.IsAbs(migrationsPath) {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		migrationsPath = filepath.Join(cwd, migrationsPath)
	}

	// Run migrations
	if err := db.RunMigrations(migrationsPath); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info().Msg("Database migrations completed successfully")
	return nil
}

// runImport performs the TinyDB import operation
func runImport(ctx context.Context, db *sqlite.DB, sourcePath string) error {
	if sourcePath == "" {
		return fmt.Errorf("source file path is required (use -source flag)")
	}

	// Check if source file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", sourcePath)
	}

	log.Info().Str("source", sourcePath).Msg("Starting TinyDB import")

	// Create repositories
	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)

	// Create importer
	importer := migration.NewImporter(activityRepo, personRepo, attendeeRepo)

	// Import data
	stats, err := importer.ImportFile(ctx, sourcePath)
	if err != nil {
		// Print partial stats even on error
		if stats != nil {
			printImportStats(stats)
		}
		return err
	}

	// Print results
	printImportStats(stats)

	if len(stats.Errors) > 0 {
		log.Warn().Int("count", len(stats.Errors)).Msg("Import completed with errors")
		fmt.Println("\nErrors:")
		for i, errMsg := range stats.Errors {
			if i >= 10 {
				fmt.Printf("  ... and %d more errors\n", len(stats.Errors)-10)
				break
			}
			fmt.Printf("  %d. %s\n", i+1, errMsg)
		}
	} else {
		log.Info().Msg("Import completed successfully with no errors")
	}

	return nil
}

// runStats displays database statistics
func runStats(ctx context.Context, db *sqlite.DB) error {
	log.Info().Msg("Fetching database statistics")

	// Create repositories
	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)

	// Get counts
	activityCount, err := activityRepo.Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count activities: %w", err)
	}

	personCount, err := personRepo.Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count people: %w", err)
	}

	// Get all attendees to count
	attendees, err := attendeeRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get attendees: %w", err)
	}
	attendeeCount := len(attendees)

	// Get table counts for other tables
	var rolesCount, expendituresCount int
	db.QueryRow("SELECT COUNT(*) FROM roles").Scan(&rolesCount)
	db.QueryRow("SELECT COUNT(*) FROM expenditures").Scan(&expendituresCount)

	// Print stats
	fmt.Println("\n=== Database Statistics ===")
	fmt.Printf("Activities:    %d\n", activityCount)
	fmt.Printf("People:        %d\n", personCount)
	fmt.Printf("Attendees:     %d\n", attendeeCount)
	fmt.Printf("Roles:         %d\n", rolesCount)
	fmt.Printf("Expenditures:  %d\n", expendituresCount)

	// Show connection stats
	stats := db.GetStats()
	fmt.Println("\n=== Connection Pool Stats ===")
	fmt.Printf("Open connections: %d\n", stats.OpenConnections)
	fmt.Printf("In use:          %d\n", stats.InUse)
	fmt.Printf("Idle:            %d\n", stats.Idle)
	fmt.Println("=============================")

	return nil
}

// printImportStats prints import statistics
func printImportStats(stats *migration.ImportStats) {
	fmt.Println("\n=== Import Statistics ===")
	fmt.Printf("Activities:\n")
	fmt.Printf("  Imported: %d\n", stats.ActivitiesImported)
	fmt.Printf("  Failed:   %d\n", stats.ActivitiesFailed)
	fmt.Printf("People:\n")
	fmt.Printf("  Imported: %d\n", stats.PeopleImported)
	fmt.Printf("  Failed:   %d\n", stats.PeopleFailed)
	fmt.Printf("Attendees:\n")
	fmt.Printf("  Imported: %d\n", stats.AttendeesImported)
	fmt.Printf("  Failed:   %d\n", stats.AttendeesFailed)
	fmt.Println("=========================")
}

// setupLogging configures the logging system
func setupLogging(verbose bool) {
	// Set log level
	logLevel := zerolog.InfoLevel
	if verbose {
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	// Use console writer for better readability
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}
