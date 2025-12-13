package repository

import (
	"context"
	"time"

	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/shopspring/decimal"
)

// ActivityRepository defines the interface for activity data access
type ActivityRepository interface {
	// Create creates a new activity
	Create(ctx context.Context, activity *domain.Activity) error

	// GetByID retrieves an activity by ID
	GetByID(ctx context.Context, id int64) (*domain.Activity, error)

	// GetAll retrieves all activities
	GetAll(ctx context.Context) ([]*domain.Activity, error)

	// Update updates an existing activity
	Update(ctx context.Context, activity *domain.Activity) error

	// Delete deletes an activity by ID
	Delete(ctx context.Context, id int64) error

	// FindByStatus retrieves activities by status
	FindByStatus(ctx context.Context, status domain.ActivityStatus) ([]*domain.Activity, error)

	// FindByType retrieves activities by type
	FindByType(ctx context.Context, activityType domain.ActivityType) ([]*domain.Activity, error)

	// FindByDateRange retrieves activities within a date range
	FindByDateRange(ctx context.Context, start, end time.Time) ([]*domain.Activity, error)

	// FindUpcoming retrieves upcoming activities (limited)
	FindUpcoming(ctx context.Context, limit int) ([]*domain.Activity, error)

	// FindRecent retrieves recently created activities (limited)
	FindRecent(ctx context.Context, limit int) ([]*domain.Activity, error)

	// Count returns the total number of activities
	Count(ctx context.Context) (int, error)

	// CountByStatus returns the count of activities by status
	CountByStatus(ctx context.Context, status domain.ActivityStatus) (int, error)
}

// PersonRepository defines the interface for person data access
type PersonRepository interface {
	// Create creates a new person
	Create(ctx context.Context, person *domain.Person) error

	// GetByID retrieves a person by ID
	GetByID(ctx context.Context, id int64) (*domain.Person, error)

	// GetAll retrieves all people
	GetAll(ctx context.Context) ([]*domain.Person, error)

	// Update updates an existing person
	Update(ctx context.Context, person *domain.Person) error

	// Delete deletes a person by ID
	Delete(ctx context.Context, id int64) error

	// FindByEmail retrieves a person by email
	FindByEmail(ctx context.Context, email string) (*domain.Person, error)

	// FindByName searches for people by name (first or last)
	FindByName(ctx context.Context, name string) ([]*domain.Person, error)

	// Search searches for people by query (name or email)
	Search(ctx context.Context, query string) ([]*domain.Person, error)

	// Count returns the total number of people
	Count(ctx context.Context) (int, error)

	// ExistsByEmail checks if a person with the given email exists
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// AttendeeRepository defines the interface for attendee data access
type AttendeeRepository interface {
	// Create creates a new attendee registration
	Create(ctx context.Context, attendee *domain.Attendee) error

	// GetByID retrieves an attendee by ID
	GetByID(ctx context.Context, id int64) (*domain.Attendee, error)

	// GetAll retrieves all attendees
	GetAll(ctx context.Context) ([]*domain.Attendee, error)

	// Update updates an existing attendee
	Update(ctx context.Context, attendee *domain.Attendee) error

	// Delete deletes an attendee by ID
	Delete(ctx context.Context, id int64) error

	// GetByActivity retrieves all attendees for an activity
	GetByActivity(ctx context.Context, activityID int64) ([]*domain.Attendee, error)

	// GetByPerson retrieves all registrations for a person
	GetByPerson(ctx context.Context, personID int64) ([]*domain.Attendee, error)

	// IsRegistered checks if a person is registered for an activity
	IsRegistered(ctx context.Context, activityID, personID int64) (bool, error)

	// CountByActivity returns the number of attendees for an activity
	CountByActivity(ctx context.Context, activityID int64) (int, error)

	// CountByPaymentStatus returns the count of attendees by payment status for an activity
	CountByPaymentStatus(ctx context.Context, activityID int64, status domain.PaymentStatus) (int, error)

	// GetPaymentSummary retrieves payment summary for an activity
	GetPaymentSummary(ctx context.Context, activityID int64) (*PaymentSummary, error)

	// FindByPaymentStatus retrieves attendees by payment status for an activity
	FindByPaymentStatus(ctx context.Context, activityID int64, status domain.PaymentStatus) ([]*domain.Attendee, error)
}

// RoleRepository defines the interface for role data access
type RoleRepository interface {
	// Create creates a new role
	Create(ctx context.Context, role *domain.Role) error

	// GetByID retrieves a role by ID
	GetByID(ctx context.Context, id int64) (*domain.Role, error)

	// GetByName retrieves a role by name
	GetByName(ctx context.Context, name string) (*domain.Role, error)

	// GetAll retrieves all roles
	GetAll(ctx context.Context) ([]*domain.Role, error)

	// GetActive retrieves all active roles
	GetActive(ctx context.Context) ([]*domain.Role, error)

	// Update updates an existing role
	Update(ctx context.Context, role *domain.Role) error

	// Delete deletes a role by ID
	Delete(ctx context.Context, id int64) error

	// ExistsByName checks if a role with the given name exists
	ExistsByName(ctx context.Context, name string) (bool, error)
}

// PaymentSummary represents payment statistics for an activity
type PaymentSummary struct {
	TotalAttendees int
	PaidCount      int
	UnpaidCount    int
	WaivedCount    int
	TotalAmount    decimal.Decimal
	PaidAmount     decimal.Decimal
	UnpaidAmount   decimal.Decimal
}

// Transactor defines the interface for transaction management
type Transactor interface {
	// WithTransaction executes a function within a transaction
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
