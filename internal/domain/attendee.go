package domain

import (
	"time"

	"github.com/PodioSpaz/event-tracker-go/internal/util"
	"github.com/shopspring/decimal"
)

// Attendee represents a person's registration for an activity
type Attendee struct {
	ID               int64
	ActivityID       int64
	PersonID         int64
	Role             AttendeeRole
	PaymentStatus    PaymentStatus
	PaymentAmount    decimal.Decimal
	PaymentDate      *time.Time
	RegistrationDate time.Time
	Notes            string
}

// NewAttendee creates a new attendee registration
func NewAttendee(activityID, personID int64, role AttendeeRole) *Attendee {
	return &Attendee{
		ActivityID:       activityID,
		PersonID:         personID,
		Role:             role,
		PaymentStatus:    PaymentStatusUnpaid,
		PaymentAmount:    decimal.Zero,
		RegistrationDate: time.Now(),
	}
}

// IsPaid returns true if the payment is marked as paid or waived
func (a *Attendee) IsPaid() bool {
	return a.PaymentStatus.IsPaid()
}

// IsUnpaid returns true if the payment is unpaid
func (a *Attendee) IsUnpaid() bool {
	return a.PaymentStatus == PaymentStatusUnpaid
}

// IsWaived returns true if the payment was waived
func (a *Attendee) IsWaived() bool {
	return a.PaymentStatus == PaymentStatusWaived
}

// HasPaymentDate returns true if a payment date is set
func (a *Attendee) HasPaymentDate() bool {
	return a.PaymentDate != nil && !a.PaymentDate.IsZero()
}

// GetRoleDisplay returns the display name for the attendee's role
func (a *Attendee) GetRoleDisplay() string {
	return a.Role.DisplayName()
}

// MarkPaid marks the attendee as paid with the given amount and date
func (a *Attendee) MarkPaid(amount decimal.Decimal, date time.Time) error {
	if amount.IsNegative() {
		return util.ErrAttendeePaymentNegative
	}

	a.PaymentStatus = PaymentStatusPaid
	a.PaymentAmount = amount
	a.PaymentDate = &date

	return nil
}

// MarkPaidNow marks the attendee as paid with the current timestamp
func (a *Attendee) MarkPaidNow(amount decimal.Decimal) error {
	return a.MarkPaid(amount, time.Now())
}

// MarkUnpaid marks the attendee as unpaid
func (a *Attendee) MarkUnpaid() {
	a.PaymentStatus = PaymentStatusUnpaid
	a.PaymentAmount = decimal.Zero
	a.PaymentDate = nil
}

// WaivePayment marks the attendee's payment as waived
func (a *Attendee) WaivePayment(date time.Time) {
	a.PaymentStatus = PaymentStatusWaived
	a.PaymentAmount = decimal.Zero
	a.PaymentDate = &date
}

// WaivePaymentNow marks the payment as waived with the current timestamp
func (a *Attendee) WaivePaymentNow() {
	a.WaivePayment(time.Now())
}

// SetPaymentAmount updates the payment amount
func (a *Attendee) SetPaymentAmount(amount decimal.Decimal) error {
	if amount.IsNegative() {
		return util.ErrAttendeePaymentNegative
	}

	a.PaymentAmount = amount
	return nil
}

// SetRole updates the attendee's role
func (a *Attendee) SetRole(role AttendeeRole) error {
	if !role.IsValid() {
		return util.ErrAttendeeRoleInvalid
	}

	a.Role = role
	return nil
}

// AddNotes adds or updates notes for the attendee
func (a *Attendee) AddNotes(notes string) {
	a.Notes = notes
}

// DaysSinceRegistration returns the number of days since registration
func (a *Attendee) DaysSinceRegistration() int {
	duration := time.Since(a.RegistrationDate)
	return int(duration.Hours() / 24)
}

// DaysSincePayment returns the number of days since payment (0 if not paid)
func (a *Attendee) DaysSincePayment() int {
	if !a.HasPaymentDate() {
		return 0
	}

	duration := time.Since(*a.PaymentDate)
	return int(duration.Hours() / 24)
}

// Validate validates the attendee data
func (a *Attendee) Validate() error {
	// Required fields
	if a.ActivityID == 0 {
		return util.ErrAttendeeActivityRequired
	}

	if a.PersonID == 0 {
		return util.ErrAttendeePersonRequired
	}

	// Validate role
	if !a.Role.IsValid() {
		return util.ErrAttendeeRoleInvalid
	}

	// Validate payment status
	if !a.PaymentStatus.IsValid() {
		return util.ErrAttendeePaymentStatusInvalid
	}

	// Validate payment amount
	if err := util.ValidateDecimal(a.PaymentAmount, "payment_amount", false); err != nil {
		return err
	}

	// Validate notes length
	if a.Notes != "" {
		if err := util.ValidateStringLength(a.Notes, "notes", 0, 1000); err != nil {
			return err
		}
	}

	// Business rules
	// If marked as paid, must have payment amount and date
	if a.PaymentStatus == PaymentStatusPaid {
		if a.PaymentAmount.IsZero() {
			return util.NewValidationError("payment_amount", "must be greater than zero when paid", util.ErrAttendeePaymentNegative)
		}

		if !a.HasPaymentDate() {
			return util.NewValidationError("payment_date", "is required when marked as paid", util.ErrRequiredField)
		}
	}

	// If waived, amount should be zero
	if a.PaymentStatus == PaymentStatusWaived && !a.PaymentAmount.IsZero() {
		return util.NewValidationError("payment_amount", "should be zero when waived", util.ErrInvalidInput)
	}

	return nil
}

// ValidatePaymentAmount checks if the payment amount matches the activity fee
func (a *Attendee) ValidatePaymentAmount(activityFee decimal.Decimal) error {
	if a.PaymentStatus == PaymentStatusWaived {
		return nil // Waived payments don't need to match
	}

	if a.PaymentStatus == PaymentStatusPaid && !a.PaymentAmount.Equal(activityFee) {
		return util.ErrPaymentAmountMismatch
	}

	return nil
}

// CanChangePaymentStatus returns true if the payment status can be changed
func (a *Attendee) CanChangePaymentStatus() bool {
	// Can always change payment status
	// (in future, might add restrictions based on business rules)
	return true
}
