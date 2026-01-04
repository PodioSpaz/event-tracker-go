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

// AttendeeRepository implements repository.AttendeeRepository for SQLite
type AttendeeRepository struct {
	db *DB
}

// NewAttendeeRepository creates a new AttendeeRepository
func NewAttendeeRepository(db *DB) repository.AttendeeRepository {
	return &AttendeeRepository{db: db}
}

// Create creates a new attendee registration
func (r *AttendeeRepository) Create(ctx context.Context, attendee *domain.Attendee) error {
	if err := attendee.Validate(); err != nil {
		return err
	}

	// Check for duplicate registration
	exists, err := r.IsRegistered(ctx, attendee.ActivityID, attendee.PersonID)
	if err != nil {
		return err
	}
	if exists {
		return util.ErrAttendeeAlreadyRegistered
	}

	query := `
		INSERT INTO attendees (
			activity_id, person_id, role, payment_status, payment_amount,
			payment_date, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	var paymentDateStr *string
	if attendee.PaymentDate != nil {
		formatted := attendee.PaymentDate.Format("2006-01-02")
		paymentDateStr = &formatted
	}

	result, err := r.db.ExecContext(ctx, query,
		attendee.ActivityID,
		attendee.PersonID,
		attendee.Role,
		attendee.PaymentStatus,
		util.FormatDecimal(attendee.PaymentAmount),
		paymentDateStr,
		attendee.Notes,
	)

	if err != nil {
		return fmt.Errorf("failed to create attendee: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	attendee.ID = id
	attendee.RegistrationDate = time.Now()

	return nil
}

// GetByID retrieves an attendee by ID
func (r *AttendeeRepository) GetByID(ctx context.Context, id int64) (*domain.Attendee, error) {
	query := `
		SELECT id, activity_id, person_id, role, payment_status, payment_amount,
		       payment_date, registration_date, notes
		FROM attendees
		WHERE id = ?
	`

	attendee := &domain.Attendee{}
	var paymentAmountStr string
	var paymentDateStr sql.NullString
	var registrationDateStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&attendee.ID,
		&attendee.ActivityID,
		&attendee.PersonID,
		&attendee.Role,
		&attendee.PaymentStatus,
		&paymentAmountStr,
		&paymentDateStr,
		&registrationDateStr,
		&attendee.Notes,
	)

	if err == sql.ErrNoRows {
		return nil, util.ErrAttendeeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get attendee: %w", err)
	}

	// Parse payment amount
	attendee.PaymentAmount, err = decimal.NewFromString(paymentAmountStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse payment amount: %w", err)
	}

	// Parse payment date
	if paymentDateStr.Valid {
		paymentDate, err := time.Parse("2006-01-02", paymentDateStr.String)
		if err == nil {
			attendee.PaymentDate = &paymentDate
		}
	}

	// Parse registration date
	attendee.RegistrationDate, _ = time.Parse("2006-01-02 15:04:05", registrationDateStr)

	return attendee, nil
}

// GetAll retrieves all attendees
func (r *AttendeeRepository) GetAll(ctx context.Context) ([]*domain.Attendee, error) {
	query := `
		SELECT id, activity_id, person_id, role, payment_status, payment_amount,
		       payment_date, registration_date, notes
		FROM attendees
		ORDER BY registration_date DESC
	`

	return r.queryAttendees(ctx, query)
}

// Update updates an existing attendee
func (r *AttendeeRepository) Update(ctx context.Context, attendee *domain.Attendee) error {
	if err := attendee.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE attendees
		SET role = ?, payment_status = ?, payment_amount = ?, payment_date = ?, notes = ?
		WHERE id = ?
	`

	var paymentDateStr *string
	if attendee.PaymentDate != nil {
		formatted := attendee.PaymentDate.Format("2006-01-02")
		paymentDateStr = &formatted
	}

	result, err := r.db.ExecContext(ctx, query,
		attendee.Role,
		attendee.PaymentStatus,
		util.FormatDecimal(attendee.PaymentAmount),
		paymentDateStr,
		attendee.Notes,
		attendee.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update attendee: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return util.ErrAttendeeNotFound
	}

	return nil
}

// Delete deletes an attendee by ID
func (r *AttendeeRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM attendees WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete attendee: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return util.ErrAttendeeNotFound
	}

	return nil
}

// GetByActivity retrieves all attendees for an activity
func (r *AttendeeRepository) GetByActivity(ctx context.Context, activityID int64) ([]*domain.Attendee, error) {
	query := `
		SELECT id, activity_id, person_id, role, payment_status, payment_amount,
		       payment_date, registration_date, notes
		FROM attendees
		WHERE activity_id = ?
		ORDER BY registration_date ASC
	`

	return r.queryAttendees(ctx, query, activityID)
}

// GetByPerson retrieves all registrations for a person
func (r *AttendeeRepository) GetByPerson(ctx context.Context, personID int64) ([]*domain.Attendee, error) {
	query := `
		SELECT id, activity_id, person_id, role, payment_status, payment_amount,
		       payment_date, registration_date, notes
		FROM attendees
		WHERE person_id = ?
		ORDER BY registration_date DESC
	`

	return r.queryAttendees(ctx, query, personID)
}

// IsRegistered checks if a person is registered for an activity
func (r *AttendeeRepository) IsRegistered(ctx context.Context, activityID, personID int64) (bool, error) {
	query := `SELECT COUNT(*) FROM attendees WHERE activity_id = ? AND person_id = ?`

	var count int
	err := r.db.QueryRowContext(ctx, query, activityID, personID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check registration: %w", err)
	}

	return count > 0, nil
}

// CountByActivity returns the number of attendees for an activity
func (r *AttendeeRepository) CountByActivity(ctx context.Context, activityID int64) (int, error) {
	query := `SELECT COUNT(*) FROM attendees WHERE activity_id = ?`

	var count int
	err := r.db.QueryRowContext(ctx, query, activityID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count attendees: %w", err)
	}

	return count, nil
}

// CountByPaymentStatus returns the count of attendees by payment status for an activity
func (r *AttendeeRepository) CountByPaymentStatus(ctx context.Context, activityID int64, status domain.PaymentStatus) (int, error) {
	query := `SELECT COUNT(*) FROM attendees WHERE activity_id = ? AND payment_status = ?`

	var count int
	err := r.db.QueryRowContext(ctx, query, activityID, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count attendees by payment status: %w", err)
	}

	return count, nil
}

// GetPaymentSummary retrieves payment summary for an activity
func (r *AttendeeRepository) GetPaymentSummary(ctx context.Context, activityID int64) (*repository.PaymentSummary, error) {
	query := `
		SELECT
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN payment_status = 'paid' THEN 1 ELSE 0 END), 0) as paid_count,
			COALESCE(SUM(CASE WHEN payment_status = 'unpaid' THEN 1 ELSE 0 END), 0) as unpaid_count,
			COALESCE(SUM(CASE WHEN payment_status = 'waived' THEN 1 ELSE 0 END), 0) as waived_count,
			COALESCE(SUM(CASE WHEN payment_status = 'paid' THEN CAST(payment_amount AS REAL) ELSE 0 END), 0) as paid_amount,
			COALESCE(SUM(CASE WHEN payment_status = 'unpaid' THEN CAST(payment_amount AS REAL) ELSE 0 END), 0) as unpaid_amount
		FROM attendees
		WHERE activity_id = ?
	`

	summary := &repository.PaymentSummary{}
	var paidAmount, unpaidAmount float64

	err := r.db.QueryRowContext(ctx, query, activityID).Scan(
		&summary.TotalAttendees,
		&summary.PaidCount,
		&summary.UnpaidCount,
		&summary.WaivedCount,
		&paidAmount,
		&unpaidAmount,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get payment summary: %w", err)
	}

	summary.PaidAmount = decimal.NewFromFloat(paidAmount).Round(2)
	summary.UnpaidAmount = decimal.NewFromFloat(unpaidAmount).Round(2)
	summary.TotalAmount = summary.PaidAmount.Add(summary.UnpaidAmount)

	return summary, nil
}

// FindByPaymentStatus retrieves attendees by payment status for an activity
func (r *AttendeeRepository) FindByPaymentStatus(ctx context.Context, activityID int64, status domain.PaymentStatus) ([]*domain.Attendee, error) {
	query := `
		SELECT id, activity_id, person_id, role, payment_status, payment_amount,
		       payment_date, registration_date, notes
		FROM attendees
		WHERE activity_id = ? AND payment_status = ?
		ORDER BY registration_date ASC
	`

	return r.queryAttendees(ctx, query, activityID, status)
}

// queryAttendees is a helper method to query attendees with optional parameters
func (r *AttendeeRepository) queryAttendees(ctx context.Context, query string, args ...interface{}) ([]*domain.Attendee, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query attendees: %w", err)
	}
	defer rows.Close()

	var attendees []*domain.Attendee

	for rows.Next() {
		attendee := &domain.Attendee{}
		var paymentAmountStr string
		var paymentDateStr sql.NullString
		var registrationDateStr string

		err := rows.Scan(
			&attendee.ID,
			&attendee.ActivityID,
			&attendee.PersonID,
			&attendee.Role,
			&attendee.PaymentStatus,
			&paymentAmountStr,
			&paymentDateStr,
			&registrationDateStr,
			&attendee.Notes,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan attendee: %w", err)
		}

		// Parse payment amount
		attendee.PaymentAmount, err = decimal.NewFromString(paymentAmountStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse payment amount: %w", err)
		}

		// Parse payment date
		if paymentDateStr.Valid {
			paymentDate, err := time.Parse("2006-01-02", paymentDateStr.String)
			if err == nil {
				attendee.PaymentDate = &paymentDate
			}
		}

		// Parse registration date
		attendee.RegistrationDate, _ = time.Parse("2006-01-02 15:04:05", registrationDateStr)

		attendees = append(attendees, attendee)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attendees: %w", err)
	}

	return attendees, nil
}
