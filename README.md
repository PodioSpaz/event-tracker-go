# Event Tracker (Go)

A desktop application for managing events, people, attendees, and expenditures. This is a Go port of the original Python Event Tracker application, built with Fyne (GUI) and SQLite (database).

## Status

**Current Version:** 0.1.0 (Released 2024-12-13)

### Completed Phases

- ✅ **Phase 1: Foundation** - Project skeleton, database connectivity, configuration
- ✅ **Phase 2: Domain Models** - Core business entities with validation
- ✅ **Phase 3: Repository Layer** - Data access layer with SQLite
- ✅ **Phase 4: Business Services** - Service layer and controllers
- ✅ **Phase 5: Migration Tool** - TinyDB import tool
- ✅ **Phase 6: GUI Foundation** - Fyne-based desktop interface
- ✅ **Phase 10: Polish & Release** - Testing, documentation, packaging, v0.1.0 release

### Upcoming Phases

- Phase 7-9: Advanced GUI Features (Forms, CSV import, enhanced UI)
- v0.2.0: Gathering support, expenditures, reports

## Documentation

- **[CHANGELOG](CHANGELOG.md)** - Version history and release notes
- **[MIGRATION GUIDE](docs/MIGRATION.md)** - Migrate from Python Event Tracker
- **[README](README.md)** - This file (project overview and usage)

## Features (MVP v0.1.0)

### Planned Features

- **Activity Management** - Create, edit, and manage events
- **People Management** - Contact database with CSV import
- **Attendee Registration** - Role-based registration with payment tracking
- **Payment Tracking** - Track payment status (paid/unpaid/waived)
- **Dashboard** - Summary statistics and upcoming events
- **Data Migration** - Import from existing Python Event Tracker

### Deferred to v0.2.0+

- Gathering support (activity type)
- Expenditure tracking
- PDF reports
- Advanced analytics

## Technology Stack

- **Language:** Go 1.23
- **GUI:** Fyne v2.5.3
- **Database:** SQLite (go-sqlite3)
- **Configuration:** Viper
- **Logging:** Zerolog
- **Decimal Precision:** shopspring/decimal

## Project Structure

```
event-tracker-go/
├── cmd/
│   ├── app/           # Main GUI application
│   └── migrate/       # Migration CLI tool
├── internal/
│   ├── domain/        # Core models (Activity, Person, Attendee)
│   ├── repository/    # Data access layer
│   │   └── sqlite/    # SQLite implementations
│   ├── service/       # Business logic
│   ├── controller/    # Application controllers
│   ├── ui/            # Fyne GUI components
│   ├── config/        # Configuration management
│   └── util/          # Utilities (validation, errors)
├── pkg/
│   └── migration/     # TinyDB importer
├── migrations/        # SQL migration files
├── data/              # Database files
├── logs/              # Log files
└── Makefile           # Build automation
```

## Getting Started

### Prerequisites

- Go 1.23 or later
- macOS, Linux, or Windows
- SQLite 3 (included with go-sqlite3)

### Installation

1. Clone the repository:
```fish
git clone https://github.com/PodioSpaz/event-tracker-go.git
cd event-tracker-go
```

2. Install dependencies:
```fish
make deps
```

3. Run database migrations:
```fish
make migrate
```

4. Build the application:
```fish
make build
```

### Running the Application

```fish
# Run the GUI application
./bin/event-tracker

# Or use make
make run

# The application will:
# 1. Load configuration
# 2. Connect to the database (creates it if needed)
# 3. Run database migrations automatically
# 4. Launch the Fyne GUI window

# Import sample data (optional, for testing)
./bin/migrate -cmd import -source testdata/sample_tinydb.json
```

**GUI Features:**
- **Dashboard** - View statistics and upcoming/recent activities
- **Activities** - Browse all activities in a table view
- **People** - List all people in the database
- **Activity Details** - Click any activity to see details and attendees
- **Person Details** - Click any person to see their registrations

### Database Management

```fish
# Run migrations
make migrate

# Show database statistics
./bin/migrate -cmd stats

# Reset database (WARNING: deletes all data)
make db-reset

# Backup database
make db-backup
```

## Development

### Available Make Targets

```fish
make help           # Show all available targets
make build          # Build application binaries
make test           # Run tests with coverage
make run            # Run the application
make migrate        # Run database migrations
make clean          # Clean build artifacts
make deps           # Download dependencies
make fmt            # Format code
make vet            # Run go vet
make lint           # Run formatters and linters
make build-all      # Build for all platforms
make bundle-macos   # Create macOS app bundle
```

### Running Tests

```fish
# Run all tests with coverage
make test

# Run tests without coverage
make test-short
```

### Database Schema

The application uses SQLite with the following tables:

- **activities** - Events and gatherings
- **people** - Contact database
- **attendees** - Registration records (junction table)
- **roles** - Attendee roles (participant, volunteer, worship_team, workshop_leader)
- **expenditures** - Expense tracking

See `migrations/001_initial_schema.sql` for the complete schema.

## Configuration

Configuration is managed through:

1. **Config file** (`config.yaml` - optional)
2. **Environment variables** (prefixed with `EVENT_TRACKER_`)
3. **Command-line flags** (where applicable)

### Default Configuration

```yaml
database:
  path: data/events.db

logging:
  level: info  # Options: debug, info, warn, error
  format: console  # Options: console, json

app:
  name: Event Tracker
  version: 0.1.0
  data_dir: data
  logs_dir: logs
```

### Environment Variables

- `EVENT_TRACKER_DATABASE_PATH` - Database file path
- `EVENT_TRACKER_LOGGING_LEVEL` - Log level (debug, info, warn, error)
- `EVENT_TRACKER_LOGGING_FORMAT` - Log format (console, json)

## Migration from Python Version

The migration tool imports data from the Python Event Tracker (TinyDB JSON format):

```fish
# Import from TinyDB JSON file
./bin/migrate -cmd import -source ../event-tracker/data/events.json

# Or specify a different database
./bin/migrate -cmd import -source ../event-tracker/data/events.json -db data/events.db

# Enable verbose logging
./bin/migrate -cmd import -source data.json -verbose
```

The tool will:
1. Automatically run database migrations
2. Parse the TinyDB JSON file
3. Validate all data
4. Import activities, people, and attendees
5. Handle duplicate emails gracefully
6. Map TinyDB IDs to SQLite IDs
7. Provide detailed import statistics

Sample import output:
```
=== Import Statistics ===
Activities:
  Imported: 15
  Failed:   0
People:
  Imported: 42
  Failed:   0
Attendees:
  Imported: 87
  Failed:   0
=========================
```

## Project Roadmap

### Phase 1: Foundation ✅ COMPLETE

- [x] Go module initialization
- [x] Directory structure
- [x] Database connection
- [x] Configuration management
- [x] Logging setup
- [x] Build system (Makefile)
- [x] Migration runner

### Phase 2: Domain Models ✅ COMPLETE

- [x] Activity model with validation
- [x] Person model with email/phone validation
- [x] Attendee model with payment logic
- [x] Role model
- [x] Custom error types
- [x] Validation utilities
- [x] Unit tests

### Phase 3: Repository Layer ✅ COMPLETE

- [x] Repository interfaces
- [x] SQLite implementations
- [x] Transaction support
- [x] Integration tests

### Phase 4: Business Services ✅ COMPLETE

- [x] Registration service - Attendee registration and management
- [x] Payment service - Payment tracking and bulk operations
- [x] Capacity service - Availability checking and capacity management
- [x] Business rules validation - Domain-level validation
- [x] Controllers - Activity and Person controllers
- [x] Service tests - Comprehensive test coverage

**Note:** Transaction support in services is implemented but requires repository-level transaction context passing for full atomic operations. This will be enhanced in a future phase.

### Phase 5: Migration Tool ✅ COMPLETE

- [x] TinyDB JSON parser - Parse Python TinyDB format
- [x] Data mappers - Convert TinyDB structures to domain models
- [x] Import service - Orchestrate import with validation
- [x] CLI interface - User-friendly command-line tool
- [x] Data validation - Comprehensive validation and error reporting
- [x] ID mapping - Map TinyDB document IDs to SQLite IDs
- [x] Duplicate handling - Skip duplicate emails gracefully
- [x] Tests - Unit tests for parser and validation

### Phase 6: GUI Foundation ✅ COMPLETE

- [x] Main window with Fyne framework
- [x] Navigation sidebar with view switching
- [x] Dashboard view with statistics cards
- [x] Upcoming and recent activities display
- [x] Activities list view with table
- [x] Activity detail view with attendees
- [x] People list view
- [x] Person detail view with registrations
- [x] Responsive layout with scrolling
- [x] Integration with controllers and services

**Features:**
- Desktop application with native look and feel
- Clean navigation between views
- Real-time data from SQLite database
- Activity and people management interfaces
- Dashboard with key statistics

### Phase 7-9: Advanced GUI Features (Future)

- [ ] Create/Edit forms for activities and people
- [ ] CSV import for people
- [ ] Payment tracking UI
- [ ] Attendee registration interface
- [ ] Search and filtering
- [ ] Reports and analytics

### Phase 10: Polish & Release

- [ ] End-to-end testing
- [ ] Documentation
- [ ] macOS app bundle
- [ ] Cross-platform builds
- [ ] v0.1.0 release

## Contributing

This is currently a personal project. Contributions, issues, and feature requests are welcome!

## License

TBD

## Acknowledgments

- Original Python Event Tracker application
- Fyne GUI framework
- Go SQLite3 driver

---

**Built with ❤️ using Go and Fyne**
