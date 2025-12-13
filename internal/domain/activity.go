package domain

import (
	"time"

	"github.com/PodioSpaz/event-tracker-go/internal/util"
	"github.com/shopspring/decimal"
)

// Activity represents an event or gathering
type Activity struct {
	ID                   int64
	Name                 string
	Description          string
	Date                 time.Time
	Location             string
	ActivityType         ActivityType
	Status               ActivityStatus
	RequiresRegistration bool
	IsFree               bool
	Fee                  decimal.Decimal
	MaxCapacity          *int // Nullable for unlimited capacity
	EstimatedHeadCount   *int // For gatherings
	ActualHeadCount      *int // For gatherings
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// NewActivity creates a new Activity with default values
func NewActivity(name, location string, date time.Time) *Activity {
	return &Activity{
		Name:                 name,
		Location:             location,
		Date:                 date,
		ActivityType:         ActivityTypeEvent,
		Status:               StatusActive,
		RequiresRegistration: true,
		IsFree:               false,
		Fee:                  decimal.Zero,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

// IsEvent returns true if this is an event
func (a *Activity) IsEvent() bool {
	return a.ActivityType == ActivityTypeEvent
}

// IsGathering returns true if this is a gathering
func (a *Activity) IsGathering() bool {
	return a.ActivityType == ActivityTypeGathering
}

// IsActive returns true if the activity is in active status
func (a *Activity) IsActive() bool {
	return a.Status == StatusActive
}

// IsCancelled returns true if the activity is cancelled
func (a *Activity) IsCancelled() bool {
	return a.Status == StatusCancelled
}

// IsCompleted returns true if the activity is completed
func (a *Activity) IsCompleted() bool {
	return a.Status == StatusCompleted
}

// CanRegisterAttendees returns true if the activity can accept registrations
func (a *Activity) CanRegisterAttendees() bool {
	return a.IsEvent() && a.RequiresRegistration && a.IsActive()
}

// IsFull returns true if the activity is at capacity
func (a *Activity) IsFull(currentAttendees int) bool {
	if !a.IsEvent() || a.MaxCapacity == nil {
		return false
	}
	return currentAttendees >= *a.MaxCapacity
}

// GetCapacityRemaining returns the number of available spots
func (a *Activity) GetCapacityRemaining(currentAttendees int) int {
	if a.MaxCapacity == nil {
		return -1 // Unlimited
	}
	remaining := *a.MaxCapacity - currentAttendees
	if remaining < 0 {
		return 0
	}
	return remaining
}

// HasCapacityFor returns true if the activity can accept N more attendees
func (a *Activity) HasCapacityFor(currentAttendees, additionalAttendees int) bool {
	if a.MaxCapacity == nil {
		return true // Unlimited capacity
	}
	return (currentAttendees + additionalAttendees) <= *a.MaxCapacity
}

// IsPast returns true if the activity date is in the past
func (a *Activity) IsPast() bool {
	return a.Date.Before(time.Now())
}

// IsFuture returns true if the activity date is in the future
func (a *Activity) IsFuture() bool {
	return a.Date.After(time.Now())
}

// DaysUntil returns the number of days until the activity
func (a *Activity) DaysUntil() int {
	duration := time.Until(a.Date)
	return int(duration.Hours() / 24)
}

// Cancel marks the activity as cancelled
func (a *Activity) Cancel() error {
	if a.IsCompleted() {
		return util.ErrActivityCompleted
	}
	a.Status = StatusCancelled
	a.UpdatedAt = time.Now()
	return nil
}

// Complete marks the activity as completed
func (a *Activity) Complete() error {
	if a.IsCancelled() {
		return util.ErrActivityCancelled
	}
	a.Status = StatusCompleted
	a.UpdatedAt = time.Now()
	return nil
}

// Reactivate marks a cancelled activity as active
func (a *Activity) Reactivate() error {
	if a.IsCompleted() {
		return util.ErrActivityCompleted
	}
	a.Status = StatusActive
	a.UpdatedAt = time.Now()
	return nil
}

// SetFee sets the activity fee and updates the IsFree flag
func (a *Activity) SetFee(fee decimal.Decimal) error {
	if fee.IsNegative() {
		return util.ErrActivityFeeNegative
	}
	a.Fee = fee
	a.IsFree = fee.IsZero()
	a.UpdatedAt = time.Now()
	return nil
}

// SetCapacity sets the maximum capacity (nil for unlimited)
func (a *Activity) SetCapacity(capacity *int) error {
	if capacity != nil && *capacity < 0 {
		return util.ErrActivityCapacityNegative
	}
	a.MaxCapacity = capacity
	a.UpdatedAt = time.Now()
	return nil
}

// Validate validates the activity data
func (a *Activity) Validate() error {
	// Required fields
	if err := util.ValidateRequired(a.Name, "name"); err != nil {
		return err
	}

	if err := util.ValidateRequired(a.Location, "location"); err != nil {
		return err
	}

	if err := util.ValidateDate(a.Date, "date"); err != nil {
		return err
	}

	// String length constraints
	if err := util.ValidateStringLength(a.Name, "name", 1, 200); err != nil {
		return err
	}

	if err := util.ValidateStringLength(a.Location, "location", 1, 200); err != nil {
		return err
	}

	if a.Description != "" {
		if err := util.ValidateStringLength(a.Description, "description", 0, 2000); err != nil {
			return err
		}
	}

	// Validate activity type
	if !a.ActivityType.IsValid() {
		return util.ErrActivityTypeInvalid
	}

	// Validate status
	if !a.Status.IsValid() {
		return util.ErrActivityStatusInvalid
	}

	// Validate fee
	if err := util.ValidateDecimal(a.Fee, "fee", false); err != nil {
		return err
	}

	// Validate capacity
	if a.MaxCapacity != nil && *a.MaxCapacity < 0 {
		return util.ErrActivityCapacityNegative
	}

	// Validate head counts (for gatherings)
	if a.EstimatedHeadCount != nil && *a.EstimatedHeadCount < 0 {
		return util.NewValidationError("estimated_head_count", "cannot be negative", util.ErrValueOutOfRange)
	}

	if a.ActualHeadCount != nil && *a.ActualHeadCount < 0 {
		return util.NewValidationError("actual_head_count", "cannot be negative", util.ErrValueOutOfRange)
	}

	// Business rule: Events with registration must have a fee defined
	if a.IsEvent() && a.RequiresRegistration && a.Fee.IsNegative() {
		return util.ErrActivityFeeNegative
	}

	return nil
}

// ValidateForRegistration validates that the activity can accept registrations
func (a *Activity) ValidateForRegistration(currentAttendees int) error {
	if !a.IsEvent() {
		return util.ErrCannotRegisterForGathering
	}

	if !a.RequiresRegistration {
		return util.ErrActivityNotAcceptingRegistrations
	}

	if a.IsCancelled() {
		return util.ErrActivityCancelled
	}

	if a.IsCompleted() {
		return util.ErrActivityCompleted
	}

	if a.IsFull(currentAttendees) {
		return util.ErrActivityFull
	}

	return nil
}
