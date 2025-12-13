package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/PodioSpaz/event-tracker-go/internal/repository"
	"github.com/PodioSpaz/event-tracker-go/internal/util"
	"github.com/shopspring/decimal"
)

// ActivityRepository implements repository.ActivityRepository for SQLite
type ActivityRepository struct {
	db *DB
}

// NewActivityRepository creates a new ActivityRepository
func NewActivityRepository(db *DB) repository.ActivityRepository {
	return &ActivityRepository{db: db}
}

// Create creates a new activity
func (r *ActivityRepository) Create(ctx context.Context, activity *domain.Activity) error {
	if err := activity.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO activities (
			name, description, date, location, activity_type, status,
			requires_registration, is_free, fee, max_capacity,
			estimated_head_count, actual_head_count
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		activity.Name,
		activity.Description,
		activity.Date.Format("2006-01-02"),
		activity.Location,
		activity.ActivityType,
		activity.Status,
		activity.RequiresRegistration,
		activity.IsFree,
		util.FormatDecimal(activity.Fee),
		activity.MaxCapacity,
		activity.EstimatedHeadCount,
		activity.ActualHeadCount,
	)

	if err != nil {
		return fmt.Errorf("failed to create activity: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	activity.ID = id
	activity.CreatedAt = time.Now()
	activity.UpdatedAt = time.Now()

	return nil
}

// GetByID retrieves an activity by ID
func (r *ActivityRepository) GetByID(ctx context.Context, id int64) (*domain.Activity, error) {
	query := `
		SELECT id, name, description, date, location, activity_type, status,
		       requires_registration, is_free, fee, max_capacity,
		       estimated_head_count, actual_head_count, created_at, updated_at
		FROM activities
		WHERE id = ?
	`

	activity := &domain.Activity{}
	var dateStr string
	var feeStr string
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&activity.ID,
		&activity.Name,
		&activity.Description,
		&dateStr,
		&activity.Location,
		&activity.ActivityType,
		&activity.Status,
		&activity.RequiresRegistration,
		&activity.IsFree,
		&feeStr,
		&activity.MaxCapacity,
		&activity.EstimatedHeadCount,
		&activity.ActualHeadCount,
		&createdAtStr,
		&updatedAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, util.ErrActivityNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	// Parse date (try multiple formats)
	activity.Date, err = time.Parse("2006-01-02", dateStr)
	if err != nil {
		// Try with timestamp format
		activity.Date, err = time.Parse("2006-01-02 15:04:05", dateStr)
		if err != nil {
			// Try ISO format
			activity.Date, err = time.Parse(time.RFC3339, dateStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse date: %w", err)
			}
		}
	}
	// Normalize to date only (remove time component)
	activity.Date = time.Date(activity.Date.Year(), activity.Date.Month(), activity.Date.Day(), 0, 0, 0, 0, time.UTC)

	// Parse fee
	activity.Fee, err = decimal.NewFromString(feeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse fee: %w", err)
	}

	// Parse timestamps
	activity.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	activity.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

	return activity, nil
}

// GetAll retrieves all activities
func (r *ActivityRepository) GetAll(ctx context.Context) ([]*domain.Activity, error) {
	query := `
		SELECT id, name, description, date, location, activity_type, status,
		       requires_registration, is_free, fee, max_capacity,
		       estimated_head_count, actual_head_count, created_at, updated_at
		FROM activities
		ORDER BY date DESC
	`

	return r.queryActivities(ctx, query)
}

// Update updates an existing activity
func (r *ActivityRepository) Update(ctx context.Context, activity *domain.Activity) error {
	if err := activity.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE activities
		SET name = ?, description = ?, date = ?, location = ?, activity_type = ?,
		    status = ?, requires_registration = ?, is_free = ?, fee = ?,
		    max_capacity = ?, estimated_head_count = ?, actual_head_count = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		activity.Name,
		activity.Description,
		activity.Date.Format("2006-01-02"),
		activity.Location,
		activity.ActivityType,
		activity.Status,
		activity.RequiresRegistration,
		activity.IsFree,
		util.FormatDecimal(activity.Fee),
		activity.MaxCapacity,
		activity.EstimatedHeadCount,
		activity.ActualHeadCount,
		activity.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update activity: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return util.ErrActivityNotFound
	}

	return nil
}

// Delete deletes an activity by ID
func (r *ActivityRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM activities WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete activity: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return util.ErrActivityNotFound
	}

	return nil
}

// FindByStatus retrieves activities by status
func (r *ActivityRepository) FindByStatus(ctx context.Context, status domain.ActivityStatus) ([]*domain.Activity, error) {
	query := `
		SELECT id, name, description, date, location, activity_type, status,
		       requires_registration, is_free, fee, max_capacity,
		       estimated_head_count, actual_head_count, created_at, updated_at
		FROM activities
		WHERE status = ?
		ORDER BY date DESC
	`

	return r.queryActivities(ctx, query, status)
}

// FindByType retrieves activities by type
func (r *ActivityRepository) FindByType(ctx context.Context, activityType domain.ActivityType) ([]*domain.Activity, error) {
	query := `
		SELECT id, name, description, date, location, activity_type, status,
		       requires_registration, is_free, fee, max_capacity,
		       estimated_head_count, actual_head_count, created_at, updated_at
		FROM activities
		WHERE activity_type = ?
		ORDER BY date DESC
	`

	return r.queryActivities(ctx, query, activityType)
}

// FindByDateRange retrieves activities within a date range
func (r *ActivityRepository) FindByDateRange(ctx context.Context, start, end time.Time) ([]*domain.Activity, error) {
	query := `
		SELECT id, name, description, date, location, activity_type, status,
		       requires_registration, is_free, fee, max_capacity,
		       estimated_head_count, actual_head_count, created_at, updated_at
		FROM activities
		WHERE date BETWEEN ? AND ?
		ORDER BY date ASC
	`

	return r.queryActivities(ctx, query, start.Format("2006-01-02"), end.Format("2006-01-02"))
}

// FindUpcoming retrieves upcoming activities (limited)
func (r *ActivityRepository) FindUpcoming(ctx context.Context, limit int) ([]*domain.Activity, error) {
	query := `
		SELECT id, name, description, date, location, activity_type, status,
		       requires_registration, is_free, fee, max_capacity,
		       estimated_head_count, actual_head_count, created_at, updated_at
		FROM activities
		WHERE date >= date('now') AND status = 'active'
		ORDER BY date ASC
		LIMIT ?
	`

	return r.queryActivities(ctx, query, limit)
}

// FindRecent retrieves recently created activities (limited)
func (r *ActivityRepository) FindRecent(ctx context.Context, limit int) ([]*domain.Activity, error) {
	query := `
		SELECT id, name, description, date, location, activity_type, status,
		       requires_registration, is_free, fee, max_capacity,
		       estimated_head_count, actual_head_count, created_at, updated_at
		FROM activities
		ORDER BY created_at DESC
		LIMIT ?
	`

	return r.queryActivities(ctx, query, limit)
}

// Count returns the total number of activities
func (r *ActivityRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM activities`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count activities: %w", err)
	}

	return count, nil
}

// CountByStatus returns the count of activities by status
func (r *ActivityRepository) CountByStatus(ctx context.Context, status domain.ActivityStatus) (int, error) {
	query := `SELECT COUNT(*) FROM activities WHERE status = ?`

	var count int
	err := r.db.QueryRowContext(ctx, query, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count activities by status: %w", err)
	}

	return count, nil
}

// queryActivities is a helper method to query activities with optional parameters
func (r *ActivityRepository) queryActivities(ctx context.Context, query string, args ...interface{}) ([]*domain.Activity, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query activities: %w", err)
	}
	defer rows.Close()

	var activities []*domain.Activity

	for rows.Next() {
		activity := &domain.Activity{}
		var dateStr string
		var feeStr string
		var createdAtStr, updatedAtStr string

		err := rows.Scan(
			&activity.ID,
			&activity.Name,
			&activity.Description,
			&dateStr,
			&activity.Location,
			&activity.ActivityType,
			&activity.Status,
			&activity.RequiresRegistration,
			&activity.IsFree,
			&feeStr,
			&activity.MaxCapacity,
			&activity.EstimatedHeadCount,
			&activity.ActualHeadCount,
			&createdAtStr,
			&updatedAtStr,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}

		// Parse date (try multiple formats)
		activity.Date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			// Try with timestamp format
			activity.Date, err = time.Parse("2006-01-02 15:04:05", dateStr)
			if err != nil {
				// Try ISO format
				activity.Date, err = time.Parse(time.RFC3339, dateStr)
				if err != nil {
					return nil, fmt.Errorf("failed to parse date: %w", err)
				}
			}
		}
		// Normalize to date only (remove time component)
		activity.Date = time.Date(activity.Date.Year(), activity.Date.Month(), activity.Date.Day(), 0, 0, 0, 0, time.UTC)

		// Parse fee
		activity.Fee, err = decimal.NewFromString(feeStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse fee: %w", err)
		}

		// Parse timestamps
		activity.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		activity.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

		activities = append(activities, activity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating activities: %w", err)
	}

	return activities, nil
}
