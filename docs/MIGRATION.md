# Migration Guide: Python Event Tracker → Go Event Tracker

## Overview

This guide explains how to migrate your data from the Python Event Tracker (which uses TinyDB/JSON storage) to the Go Event Tracker (which uses SQLite).

## Prerequisites

- **Python Event Tracker data file:** Located at `../event-tracker/data/events.json` (or your custom location)
- **Go Event Tracker migration tool:** Built binary at `bin/migrate`
- **Go 1.23 or later:** For building from source

## Quick Start

```fish
# 1. Backup your Python data
cp ../event-tracker/data/events.json ../event-tracker/data/events.backup.json

# 2. Build the migration tool
make build

# 3. Run the import
./bin/migrate -cmd import -source ../event-tracker/data/events.json

# 4. Verify the import
./bin/migrate -cmd stats
```

## Detailed Migration Steps

### Step 1: Backup Your Data

**CRITICAL:** Always backup your data before migrating:

```fish
# Backup Python TinyDB file
cp ../event-tracker/data/events.json ../event-tracker/data/events.backup.json

# If you have an existing Go database, backup it too
cp data/events.db data/events.backup.db 2>/dev/null || true
```

### Step 2: Build the Migration Tool

```fish
# Build both the app and migration tool
make build

# Or build just the migration tool
go build -o bin/migrate ./cmd/migrate
```

### Step 3: Run the Import

Basic import:
```fish
./bin/migrate -cmd import -source ../event-tracker/data/events.json
```

Import to a specific database:
```fish
./bin/migrate -cmd import -source ../event-tracker/data/events.json -db data/custom.db
```

Import with verbose logging:
```fish
./bin/migrate -cmd import -source ../event-tracker/data/events.json -verbose
```

### Step 4: Verify the Import

Check database statistics:
```fish
./bin/migrate -cmd stats
```

Expected output:
```
=== Database Statistics ===
Activities:    15
People:        42
Attendees:     87
Roles:         4
Expenditures:  0

=== Connection Pool Stats ===
Open connections: 1
In use:          0
Idle:            1
=============================
```

### Step 5: Launch the Application

```fish
# Run the GUI application
./bin/event-tracker

# Or use make
make run
```

## What Gets Migrated

### ✅ Migrated Data

- **People Records:**
  - First name, last name
  - Email address
  - Phone number
  - Creation/update timestamps

- **Activities (Events only):**
  - Name, description
  - Date, location
  - Activity type (converted to 'event')
  - Status (active/cancelled/completed)
  - Registration settings
  - Fee information
  - Capacity limits
  - Head count estimates

- **Attendee Registrations:**
  - Activity-Person relationships
  - Roles (participant, volunteer, worship_team, workshop_leader)
  - Payment status (paid/unpaid/waived)
  - Payment amounts and dates
  - Registration dates
  - Notes

- **Role Definitions:**
  - Standard 4 roles pre-seeded
  - Role display names and descriptions

### ❌ Not Migrated (Deferred to v0.2.0)

- **Expenditures:** Expense tracking not yet implemented
- **Gathering type activities:** Only Events supported in v0.1.0
- **Custom roles:** Beyond the standard 4 roles
- **Historical audit logs:** If any exist in Python version

## Data Transformations

### Activity Type Conversion

All activities from the Python `events` table are converted to type `'event'`:
- Python schema: Separate `events` and `gatherings` tables
- Go schema: Single `activities` table with `activity_type` field

### Date Format

- **Python TinyDB:** ISO date strings `"2024-03-15"`
- **Go SQLite:** DATE type, stored internally as Julian day

### Decimal Values

- **Python TinyDB:** JSON numbers `25.5`
- **Go SQLite:** TEXT strings `"25.50"` (using shopspring/decimal)
- Ensures precise monetary calculations

### ID Mapping

- **Python TinyDB:** Document IDs like `{"1": {...}, "2": {...}}`
- **Go SQLite:** Auto-increment INTEGER PRIMARY KEY
- Migration tool maintains internal mapping for foreign key resolution

## Troubleshooting

### Problem: "No such file or directory" error

**Solution:** Verify the source file path

```fish
# Check if file exists
ls -la ../event-tracker/data/events.json

# Use absolute path if needed
./bin/migrate -cmd import -source /full/path/to/events.json
```

### Problem: "Failed to parse TinyDB file"

**Solution:** Verify JSON format

```fish
# Validate JSON syntax
python -m json.tool ../event-tracker/data/events.json > /dev/null

# Check file contents
head -20 ../event-tracker/data/events.json
```

Expected TinyDB structure:
```json
{
  "_default": {},
  "roles": {"1": {...}, "2": {...}},
  "people": {"1": {...}, "2": {...}},
  "activities": {"1": {...}, "2": {...}},
  "event_attendees": {"1": {...}, "2": {...}}
}
```

### Problem: Duplicate email errors during import

**Behavior:** Migration tool skips duplicate emails gracefully

```
INF Skipping duplicate person email=alice@example.com
```

**Solution:** This is expected behavior. The import will:
1. Skip the duplicate person record
2. Use the existing person ID for attendee relationships
3. Continue importing remaining data
4. Report statistics at the end

### Problem: Database locked error

**Solution:** Close any other connections

```fish
# Kill any running instances
pkill -f event-tracker

# Remove lock files
rm -f data/events.db-wal data/events.db-shm

# Try import again
./bin/migrate -cmd import -source ../event-tracker/data/events.json
```

### Problem: Missing migrations directory

**Error:** `Failed to run database migrations: no such file or directory`

**Solution:** Ensure you're in the project root

```fish
# Check current directory
pwd
# Should be: /path/to/event-tracker-go

# Check migrations exist
ls -la migrations/
# Should show: 001_initial_schema.sql

# Run from project root
./bin/migrate -cmd import -source ../event-tracker/data/events.json
```

## Migration Statistics

After import, you'll see statistics like:

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

### Understanding the Numbers

- **Imported:** Successfully migrated records
- **Failed:** Records that couldn't be imported (check logs for details)
- **Skipped duplicates:** Not counted in either category (noted in logs)

## Post-Migration Checklist

- [ ] Verify record counts match Python version
- [ ] Spot-check a few activities in the GUI
- [ ] Verify attendee registrations are intact
- [ ] Check payment statuses are correct
- [ ] Test creating new activities
- [ ] Test registering new attendees
- [ ] Backup the SQLite database

## Advanced Usage

### Dry Run (Validation Only)

To validate your TinyDB file without importing:

```fish
# Not currently supported in v0.1.0
# Coming in v0.2.0
```

### Selective Import

To import only specific record types:

```fish
# Not currently supported in v0.1.0
# Coming in v0.2.0
```

### Re-running Import

⚠️ **WARNING:** Re-running import will create duplicate records!

To start fresh:
```fish
# Reset database
make db-reset

# Re-import
./bin/migrate -cmd import -source ../event-tracker/data/events.json
```

## Database Schema Reference

For details on the SQLite schema, see:
- `migrations/001_initial_schema.sql`
- README.md - Database Schema section

## Getting Help

- **Documentation:** See README.md
- **Issues:** https://github.com/PodioSpaz/event-tracker-go/issues
- **Source Code:** View migration implementation in `pkg/migration/`

## Version Compatibility

- **v0.1.0:** Supports TinyDB format from Python Event Tracker v1.x
- **Future versions:** Will support direct database-to-database migration
