# Changelog

All notable changes to Event Tracker Go will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2024-12-13

### Added
- Initial Go implementation of Event Tracker desktop application
- Core domain models: Activity, Person, Attendee, Role
- SQLite database with full CRUD operations and migration system
- Fyne-based desktop GUI with modern interface
- Dashboard view with summary statistics and recent activities
- Activities management (create, view, filter by status, detail views)
- People management (create, view, search by name/email)
- Attendee registration with role-based tracking (participant, volunteer, worship_team, workshop_leader)
- Payment tracking with three states (paid/unpaid/waived) and decimal precision
- TinyDB migration tool for importing data from Python Event Tracker
- Cross-platform build support (macOS Intel, macOS ARM64, Linux, Windows)
- Comprehensive test suite with 84+ unit and integration tests
- Embedded SQL migrations for standalone distribution
- macOS app bundle (.app) packaging
- Configuration management via file, environment variables, and defaults
- Structured logging with configurable levels (debug, info, warn, error)
- Database connection pooling and transaction support
- Email and phone validation for person records
- Capacity checking and duplicate registration prevention
- CLI migration tool with import, stats, and up commands

### Technical Details
- **Language:** Go 1.23+
- **GUI Framework:** Fyne v2.5.3
- **Database:** SQLite with go-sqlite3 driver
- **Decimal Handling:** shopspring/decimal for monetary precision
- **Configuration:** Viper for flexible configuration
- **Logging:** Zerolog for structured logging
- **Testing:** testify for assertions and mocking

### Features Deferred to v0.2.0
- Gathering support (only Events supported in v0.1.0)
- Expenditure tracking
- PDF report generation
- CSV export functionality
- Advanced analytics and reporting
- Activity type conversion
- Bulk operations

### Known Limitations
- macOS app bundle is unsigned (will show security warning on first launch)
- GUI testing is limited (TUI components difficult to test)
- No Windows/Linux app bundling (binaries only)
- Migration tool requires external migrations directory (GUI app has embedded migrations)

## Links

- [Repository](https://github.com/PodioSpaz/event-tracker-go)
- [Issues](https://github.com/PodioSpaz/event-tracker-go/issues)
